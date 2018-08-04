package docker

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"github.com/jhoonb/archivex"
	uuid "github.com/satori/go.uuid"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/util/homedir"
)

//PushImage - takes the image name as an argument in the format of
/// <org>/<name>:<tag>
func PushImage(imageName string) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	registryURL := os.Getenv("DOCKER_REGISTRY_URL")
	if username == "" || password == "" {
		return fmt.Errorf("you didnt set a DOCKER_USERNAME or DOCKER_PASSWORD")
	}

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	if registryURL != "" {
		authConfig.ServerAddress = registryURL
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	out, err := cli.ImagePush(context.Background(), imageName, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}

	defer out.Close()
	err = parseReader(out)
	if err != nil {
		return fmt.Errorf("failed reading response body: %v", err)
	}

	return nil
}

func HelmInstall(name, image, tag, port string, helmValues map[string]interface{}) error {
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	kubeConfigReader, err := os.Open(kubeConfigPath)
	if err != nil {
		return fmt.Errorf("couldnt read file: %v", err)
	}

	config, err := ioutil.TempFile(".", "config")
	defer os.Remove(config.Name())
	if err != nil {
		return fmt.Errorf("create tmp file: %v", err)
	}

	io.Copy(config, kubeConfigReader)
	helmValuesFile, err := ioutil.TempFile(".", "helm_values")
	defer os.Remove(helmValuesFile.Name())
	if err != nil {
		return fmt.Errorf("create helm values file: %v", err)
	}

	helmValuesBytes, err := yaml.Marshal(helmValues)
	if err != nil {
		return fmt.Errorf("failed getting helm yaml: %v", err)
	}

	io.Copy(helmValuesFile, bytes.NewReader(helmValuesBytes))
	r := HelmInstallDockerfile(config.Name(), helmValuesFile.Name(), image, tag, name, port)
	err = RunDockerfileInTmpImage(r)
	if err != nil {
		return fmt.Errorf("build image failed: %v", err)
	}
	return nil
}

func HelmInstallDockerfile(kubeConfigPath, valuesYamlPath, repo, tag, name, port string) io.Reader {
	fileBytes := []byte(
		fmt.Sprintf(
			dockerFileHelmInstall,
			kubeConfigPath,
			valuesYamlPath,
			name,
			repo,
			tag,
			port,
		),
	)
	r := bytes.NewReader(fileBytes)
	return r
}

// RunDockerfileInTmpImage - use a dockerfile as a script to run
// in a tmp container
func RunDockerfileInTmpImage(dockerFileReader io.Reader) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return fmt.Errorf("couldnt create client: %v", err)
	}

	guid, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed generating a guid: %v", err)
	}

	testImageName := "haikube-runner:" + guid.String()
	err = BuildImage(dockerFileReader, testImageName)
	if err != nil {
		return fmt.Errorf("build image failed: %v", err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list images: %v", err)
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == testImageName {
				imageID := image.ID
				fmt.Println("cleaning up", imageID)
				_, err := cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{
					PruneChildren: true,
					Force:         true,
				})
				if err != nil {
					return fmt.Errorf("failed to remove image: %v", err)
				}
			}
		}
	}
	return nil
}

func BuildImage(dockerFileReader io.Reader, imagename string) error {
	dockerFile, err := ioutil.TempFile(".", "Dockerfile")
	io.Copy(dockerFile, dockerFileReader)
	defer os.Remove(dockerFile.Name())
	tarname := os.TempDir() + "/bld-" + dockerFile.Name()
	tar := new(archivex.TarFile)
	tar.Create(tarname)
	tar.AddAll(".", false)
	tar.Close()
	dockerBuildContext, err := os.Open(tarname + ".tar")
	if err != nil {
		return err
	}

	defer dockerBuildContext.Close()
	defer os.Remove(tarname + ".tar")
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	imageBuildResponse, err := cli.ImageBuild(
		context.Background(),
		dockerBuildContext,
		types.ImageBuildOptions{
			Tags:        []string{imagename},
			Context:     dockerBuildContext,
			Dockerfile:  dockerFile.Name(),
			Remove:      true,
			ForceRemove: true,
			NetworkMode: "host",
			NoCache:     true,
		},
	)
	defer func() {
		if imageBuildResponse.Body != nil {
			imageBuildResponse.Body.Close()
		}
	}()
	if err != nil {
		log.Fatal("unable to build docker image: ", err)
	}

	err = parseReader(imageBuildResponse.Body)
	if err != nil {
		return fmt.Errorf("failed reading response body: %v", err)
	}

	return nil
}

func parseReader(rdr io.Reader) error {
	bodyReader := bufio.NewReader(rdr)
	for {
		line, _, err := bodyReader.ReadLine()
		if err != nil {
			break
		}

		m := struct {
			Error       interface{} `json:"error"`
			ErrorDetail interface{} `json:"errorDetail"`
			Stream      interface{} `json:"stream"`
			Aux         interface{} `json:"aux"`
		}{}
		json.Unmarshal(line, &m)
		if m.Aux != nil {
			fmt.Fprint(os.Stdout, m.Aux)
		}

		if m.Stream != nil {
			fmt.Fprint(os.Stdout, m.Stream)
		}

		if m.Error != nil {
			fmt.Fprint(os.Stdout, m.Error)
			return fmt.Errorf(fmt.Sprint(m.Error))
		}

		if m.ErrorDetail != nil {
			fmt.Fprint(os.Stdout, m.ErrorDetail)
			return fmt.Errorf(fmt.Sprint(m.ErrorDetail))
		}
	}
	return nil
}

// CreateDockerfile returns a reader which contains a dockerfiles contents
// as well as a function which can be used to cleanup any tmp
// dir created to store buildpacks when they are dloaded & unzipped
// it will also return an error if anything goes wrong
func CreateDockerfile(
	buildpackURI,
	baseImage,
	port,
	codepath,
	cmd string,
	envmap map[string]string,
	downloader func(string) (string, error),
) (io.Reader, func(), error) {
	buildpackpath, err := downloader(buildpackURI)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed downloading: %v", err)
	}

	tempBuildpackUnzipped, err := ioutil.TempDir(".", "buildpack")
	cleanBuildpackTmp := func() {
		os.Remove(buildpackpath)
		os.RemoveAll(tempBuildpackUnzipped)
	}
	if err != nil {
		return nil, cleanBuildpackTmp, fmt.Errorf("creation of tmp dir for buildpack failed: %v", err)
	}

	err = unzip(buildpackpath, tempBuildpackUnzipped)
	if err != nil {
		return nil, cleanBuildpackTmp, fmt.Errorf("unzip of buildpack failed: %v", err)
	}

	var fileBytes []byte
	if _, err := os.Stat(tempBuildpackUnzipped + "/bin/finalize"); os.IsNotExist(err) {
		fileBytes = []byte(
			fmt.Sprintf(
				dockerFileBuildpackLegacy,
				baseImage,
				codepath,
				tempBuildpackUnzipped,
				dockerFileEnvCmdFromMap(envmap),
				port,
				port,
				cmd,
			),
		)

	} else {
		fileBytes = []byte(
			fmt.Sprintf(
				dockerFileBuildpackNew,
				baseImage,
				codepath,
				tempBuildpackUnzipped,
				dockerFileEnvCmdFromMap(envmap),
				port,
				port,
				cmd,
			),
		)
	}

	r := bytes.NewReader(fileBytes)
	return r, cleanBuildpackTmp, nil
}

func dockerFileEnvCmdFromMap(envmap map[string]string) string {
	resp := ""
	for k, v := range envmap {
		resp += fmt.Sprintf("ENV %s=%s\n", k, v)
	}
	return resp
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func URIDownloader(downloadURI string) (string, error) {
	out, err := ioutil.TempFile(".", "buildpack")
	defer out.Close()
	if err != nil {
		return "", fmt.Errorf("tmp file create failed: %v", err)
	}

	resp, err := http.Get(downloadURI)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("httpGet failed for %s: %v", downloadURI, err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("copy response body failed: %v", err)
	}

	fileInfo, err := out.Stat()
	if err != nil {
		return "", fmt.Errorf("stat downloaded file failed: %v", err)
	}

	return fileInfo.Name(), nil
}

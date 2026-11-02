package main

import (
	"context"
	"time"
	"dagger/kuma-platform/internal/dagger"
)

type KumaPlatform struct{}

// +cache="never"
func (m *KumaPlatform) DeployApp(
	ctx context.Context,
	sock *dagger.Socket,
	// +defaultPath="/"
	// +ignore=[".env", ".git", "dagger"]
	src *dagger.Directory,
	hostDir string,
	envFile *dagger.Secret,
	// +optional
	galera bool,
) (string, error) {
	file := "app/docker-compose.yaml"
	if galera {
		file = "galera/docker-compose.yaml"
	}

	return dag.Container().
		From("docker:29-cli").
		WithUnixSocket("/var/run/docker.sock", sock).
		WithMountedDirectory(hostDir, src).
		WithWorkdir(hostDir).
		WithMountedSecret("/run/secrets/app.env", envFile).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c",
			"docker network inspect monitoring >/dev/null 2>&1 || docker network create monitoring && " +
				"docker compose --env-file /run/secrets/app.env -f " + file + " up -d --wait &&" +
				"docker compose --env-file /run/secrets/app.env -f " + file + " ps"}).
		Stdout(ctx)
}

// +cache="never"
func (m *KumaPlatform) DeployMonitoring(
	ctx context.Context,
	sock *dagger.Socket,
	// +defaultPath="/"
	// +ignore=[".env", ".git", "dagger"]
	src *dagger.Directory,
	hostDir string,
	envFile *dagger.Secret,
) (string, error) {
	return dag.Container().
		From("docker:29-cli").
		WithUnixSocket("/var/run/docker.sock", sock).
		WithMountedDirectory(hostDir, src).
		WithWorkdir(hostDir).
		WithMountedSecret("/run/secrets/monitoring.env", envFile).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c",
			"docker network inspect monitoring >/dev/null 2>&1 || docker network create monitoring && " +
				"docker compose --env-file /run/secrets/monitoring.env -f monitoring/docker-compose.yaml up -d --wait &&" +
				"docker compose --env-file /run/secrets/monitoring.env -f monitoring/docker-compose.yaml ps"}).
		Stdout(ctx)
}

// +cache="never"
func (m *KumaPlatform) ConfigureGrafana(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=[".env", ".git", "dagger"]
	src *dagger.Directory,
	envFile *dagger.Secret,
	grafana *dagger.Service,
) (string, error) {
	return dag.Container().
		From("curlimages/curl:8.23.0").
		WithServiceBinding("grafana", grafana).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedSecret("/run/secrets/app.env", envFile, dagger.ContainerWithMountedSecretOpts{
			Owner: "curl_user",
		}).
		WithEnvVariable("GRAFANA_URL", "http://grafana:3000").
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c",
			"set -a && . /run/secrets/app.env && set +a && sh monitoring/grafana/configure-grafana.sh"}).
		Stdout(ctx)
}

// +cache="never"
func (m *KumaPlatform) DeployAll(
	ctx context.Context,
	sock *dagger.Socket,
	// +defaultPath="/"
	// +ignore=[".env", ".git", "dagger"]
	src *dagger.Directory,
	hostDir string,
	envFile *dagger.Secret,
	grafana *dagger.Service,
	// +optional
	galera bool,
) (string, error) {
	app, err := m.DeployApp(ctx, sock, src, hostDir, envFile, galera)
	if err != nil {
		return "", err
	}

	mon, err := m.DeployMonitoring(ctx, sock, src, hostDir, envFile)
	if err != nil {
		return "", err
	}

	gr, err := m.ConfigureGrafana(ctx, src, envFile, grafana)
	if err != nil {
		return "", err
	}

	return app + "\n" + mon + "\n" + gr, nil
}

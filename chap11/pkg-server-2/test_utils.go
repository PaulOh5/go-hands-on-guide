package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
)

func getTestBucket(tmpDir string) (*blob.Bucket, error) {
	myDir, err := os.MkdirTemp(tmpDir, "test-bucket")
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(fmt.Sprintf("file:///%s", myDir))
	if err != nil {
		return nil, err
	}
	opts := fileblob.Options{
		URLSigner: fileblob.NewURLSignerHMAC(
			u, []byte("super secret"),
		),
	}
	return fileblob.OpenBucket(myDir, &opts)
}

func getTestDb() (testcontainers.Container, *sql.DB, error) {
	myDir, err := os.Stat("mysql-init")
	if err != nil {
		return nil, nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	bindMountPath := filepath.Join(cwd, myDir.Name())

	waitForSql := wait.ForSQL("3306/tcp", "mysql", func(host string, p nat.Port) string {
		return "root:rootpw@tcp(127.0.0.1:" + p.Port() + ")/pakcage_server"
	})
	waitForSql.WithPollInterval(5 * time.Second)
	waitForSql.WithStartupTimeout(1 * time.Minute)

	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_DATABASE":      "package_server",
			"MYSQL_USER":          "packages_rw",
			"MYSQL_PASSWORD":      "password",
			"MYSQL_ROOT_PASSWORD": "rootpw",
		},
		ImagePlatform: "linux/x86_64",
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      bindMountPath,
				ContainerFilePath: "/docker-entrypoint-initdb.d",
				FileMode:          0o777,
			},
		},
		WaitingFor: waitForSql,
	}
	ctx := context.Background()
	mysqlC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		return mysqlC, nil, err
	}
	addr, err := mysqlC.PortEndpoint(ctx, "3306", "")
	if err != nil {
		return mysqlC, nil, err
	}
	db, err := getDatabaseConn(
		addr, "package_server",
		"package_rw", "password",
	)
	if err != nil {
		return mysqlC, nil, nil
	}
	return mysqlC, db, nil
}

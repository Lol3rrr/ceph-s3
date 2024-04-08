package main

import (
	"context"
	"log"
	"os"

	"github.com/lol3rrr/ceph-s3/ceph"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	err := Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func Run() error {
	dbplugin.ServeMultiplex(testFactor)

	return nil
}

func testFactor() (interface{}, error) {
	return New()
}

func New() (interface{}, error) {
	logger := hclog.Default()
	logger.Error("New")

	db, err := newDatabase()
	if err != nil {
		return nil, err
	}

	// This middleware isn't strictly required, but highly recommended to prevent accidentally exposing
	// values such as passwords in error messages. An example of this is included below
	idb := dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.secretValues)
	return idb, nil
}

type MyDatabase struct {
	// Variables for the database
	username string
	password string
}

func newDatabase() (*MyDatabase, error) {
	db := &MyDatabase{
		// ...
	}
	return db, nil
}

func (db *MyDatabase) secretValues() map[string]string {
	return map[string]string{
		db.password: "[password]",
	}
}

func (db *MyDatabase) Close() error {
	return nil
}

func (db *MyDatabase) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	ceph, err := ceph.CephAuth(db.username, db.password)
	if err != nil {
		return dbplugin.DeleteUserResponse{}, err
	}

	err = ceph.RemoveKey("testing", req.Username)
	if err != nil {
		return dbplugin.DeleteUserResponse{}, err
	}

	return dbplugin.DeleteUserResponse{}, nil
}

func (db *MyDatabase) Initialize(ctx context.Context, init_req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	logger := hclog.Default()
	logger.Error("New Database")
	logger.Error("Init-Req: ", init_req)

	db.username = init_req.Config["ceph_username"].(string)
	db.password = init_req.Config["ceph_password"].(string)

	ceph, err := ceph.CephAuth(db.username, db.password)
	if err != nil {
		return dbplugin.InitializeResponse{}, err
	}

	users, err := ceph.RgwUsers()
	if err != nil {
		return dbplugin.InitializeResponse{}, err
	}

	logger.Error("Users: ", users)

	return dbplugin.InitializeResponse{}, nil
}

func (db *MyDatabase) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	ceph, err := ceph.CephAuth(db.username, db.password)
	if err != nil {
		return dbplugin.NewUserResponse{}, err
	}

	response, err := ceph.AddKey("testing", "", req.Password)
	if err != nil {
		return dbplugin.NewUserResponse{}, err
	}

	return dbplugin.NewUserResponse{
		Username: response.AccessKey,
	}, nil
}

func (db *MyDatabase) Type() (string, error) {
	return "Ceph-S3", nil
}

func (db *MyDatabase) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	return dbplugin.UpdateUserResponse{}, nil
}

//go:build integration

package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"webapp/pkg/data"
	"webapp/pkg/repository"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

// var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {

	// connect to docker;
	// fail if docker not running;
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; check its running...\nerror:%s", err)
	}

	pool = p
	// setup our docker options, specifying image etc
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	// get a resource (docker image)
	resource, err := pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// start image and wait until its ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("Error:" + err.Error())
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to database: %s", err)
	}

	//populate db with empty tables
	err = createTables()
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	// setup test repo
	testRepo = &PostgresDBRepo{DB: testDB}

	// run the tests
	code := m.Run()

	// cleanup docker
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {

	tableSql, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSql))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("can't ping database")
	}
}

func TestPostgresDBRepo_InsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("insert user returned an error: %s", err)
	}

	if id != 1 {
		t.Errorf("expected user id of 1, but got %d", id)
	}
}

func TestPostgresDBRepo_AllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("failed to list users in database: %s", err)
	}

	if len(users) != 1 {
		t.Errorf("AllUsers reported wrong size, expected 1 user but got %d", len(users))
	}

	testUser := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "Jack@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("failed to list users in database: %s", err)
	}

	if len(users) != 2 {
		t.Errorf("AllUsers reported wrong size, expected 2 user but got %d", len(users))
	}

}

func TestPostgresDBRepo_GetUser(t *testing.T) {

	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("error getting user by id: %s", err)
	}

	if user.Email != "admin@example.com" {
		t.Errorf("wrong email returned by GetUser, expected admin@example.com, got: %s", user.Email)
	}
	_, err = testRepo.GetUser(3)
	if err == nil {
		t.Error("no error reported when getting non existant user by id")
	}
}

func TestPostgresDBRepo_GetUserByEmail(t *testing.T) {

	user, err := testRepo.GetUserByEmail("Jack@example.com")
	if err != nil {
		t.Errorf("error getting user by id: %s", err)
	}

	if user.FirstName != "Jack" {
		t.Errorf("wrong name returned by GetUserByEmail, expected Jack, got: %s", user.FirstName)
	}

	_, err = testRepo.GetUserByEmail("fake@email.com")
	if err == nil {
		t.Error("no error reported when getting non existant user by email")
	}
}

func TestPostgresDBRepo_UpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(2)
	user.FirstName = "Jacky"
	user.Email = "Jacky@example.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("error updating user %d: %s", 2, err)
	}

	user, _ = testRepo.GetUser(2)

	if user.FirstName != "Jacky" {
		t.Errorf("update to user failed, expected firstname of Jacky but got %s", user.FirstName)
	}
	if user.Email != "Jacky@example.com" {
		t.Errorf("update to user failed, expected email of Jacky@example.com but got %s", user.Email)
	}
}

func TestPosgresDBRepo_DeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("error deleting user: %s", err)
	}

	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Error("no error returned when retrieving deleted user from database")
	}
}

func TestPostgresDBRepo_ResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "newPassword")
	if err != nil {
		t.Errorf("error updating user password: %s", err)
	}

	user, _ := testRepo.GetUser(1)

	matches, err := user.PasswordMatches("newPassword")
	if err != nil {
		t.Error(err)
	}

	if !matches {
		t.Error("password does not match newPassword")
	}
}

func TestPostgresDBRepo_InsertUserImage(t *testing.T) {
	image := data.UserImage{
		UserID:    1,
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	newID, err := testRepo.InsertUserImage(image)
	if err != nil {
		t.Errorf("failed to insert user image: %s", err)
	}

	if newID != 1 {
		t.Errorf("Insert Image should have ID of 1, but got %d", newID)
	}

	image.UserID = -1

	_, err = testRepo.InsertUserImage(image)
	if err == nil {
		t.Error("inserted a user image with non existant user id")
	}
}

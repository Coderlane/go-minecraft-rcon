package client

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/Coderlane/go-minecraft-rcon/client"
	"github.com/golang/mock/gomock"
)

type testContext struct {
	ctrl   *gomock.Controller
	client *client.MockClient
	mc     *MinecraftClient
}

func newTestContext(t *testing.T) *testContext {
	ctrl := gomock.NewController(t)
	client := client.NewMockClient(ctrl)
	mc := NewMinecraftClient(client)
	return &testContext{
		ctrl:   ctrl,
		client: client,
		mc:     mc,
	}
}

func (tc *testContext) Finish() {
	tc.ctrl.Finish()
}

func expectError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), contains) {
		t.Errorf("Expected \"%s\": %v", contains, err)
	}
}

func TestUsersListSuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	tc.client.EXPECT().Request("list").
		Return("There are 0 of a max of 20 players online:", nil)
	users, err := tc.mc.UsersList()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Got: %v Expected: %v", users, expected)
	}
}

func TestUsersListEmptySuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	tc.client.EXPECT().Request("list").
		Return("There are 2 of a max of 20 players online: user1, user2", nil)
	users, err := tc.mc.UsersList()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"user1", "user2"}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Got: %v Expected: %v", users, expected)
	}
}

func TestHelpSuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	tc.client.EXPECT().Request("help").
		Return("/list", nil)
	users, err := tc.mc.Help()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"list"}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Got: %v Expected: %v", users, expected)
	}
}

func TestHelpCmdSuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	tc.client.EXPECT().Request("help list").
		Return("/list", nil)
	users, err := tc.mc.HelpCmd("list")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"list"}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Got: %v Expected: %v", users, expected)
	}
}

func TestIPBanPardonSuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	ip := net.ParseIP("127.0.0.1")
	tc.client.EXPECT().Request("ban-ip 127.0.0.1").
		Return("Banned IP 127.0.0.1: Banned by an operator", nil)
	if err := tc.mc.IPBan(ip); err != nil {
		t.Fatal(err)
	}
	tc.client.EXPECT().Request("pardon-ip 127.0.0.1").
		Return("Unbanned IP 127.0.0.1", nil)
	if err := tc.mc.IPPardon(ip); err != nil {
		t.Fatal(err)
	}
}

func TestUserBanPardonSuccess(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	tc.client.EXPECT().Request("ban test").
		Return("Banned test: Banned by an operator", nil)
	if err := tc.mc.UserBan("test"); err != nil {
		t.Fatal(err)
	}
	tc.client.EXPECT().Request("pardon test").
		Return("Unbanned test", nil)
	if err := tc.mc.UserPardon("test"); err != nil {
		t.Fatal(err)
	}
}

var (
	invalidUsers []string = []string{"", "t", "test$", "aaaaaaaaaaaaaaaaa"}
)

func TestUserBanInvalidUserFails(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	for _, user := range invalidUsers {
		t.Run(user, func(t *testing.T) {
			err := tc.mc.UserBan(user)
			expectError(t, err, fmt.Sprintf("invalid user: %s", user))
		})
	}
}

func TestUserPardonInvalidUserFails(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	for _, user := range invalidUsers {
		t.Run(user, func(t *testing.T) {
			err := tc.mc.UserPardon(user)
			expectError(t, err, fmt.Sprintf("invalid user: %s", user))
		})
	}
}

func TestUserBanErrorReturned(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	tc.client.EXPECT().Request("ban test").
		Return("Nothing changed. The player is already banned", nil)
	err := tc.mc.UserBan("test")
	expectError(t, err, "Nothing changed.")
}

func TestUserPardonErrorReturned(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()
	tc.client.EXPECT().Request("pardon test").
		Return("Nothing changed. The player isn't banned", nil)
	err := tc.mc.UserPardon("test")
	expectError(t, err, "Nothing changed.")
}

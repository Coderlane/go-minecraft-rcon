package client

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/Coderlane/go-minecraft-rcon/client"
)

var (
	userRegex *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_]{3,16}$`)
)

func validateUser(user string) error {
	if !userRegex.MatchString(user) {
		return fmt.Errorf("invalid user: %s", user)
	}
	return nil
}

func validateResponsePrefix(response, expected string) error {
	if !strings.HasPrefix(response, expected) {
		return fmt.Errorf("%s", response)
	}
	return nil
}

// MinecraftClient is a high level wrapper for the Minecraft RCon API
type MinecraftClient struct {
	client client.Client
}

// NewMinecraftClient creates a new client and attaches to the server.
func NewMinecraftClient(client client.Client) *MinecraftClient {
	return &MinecraftClient{
		client: client,
	}
}

// Close closes the connection with the server.
func (mc *MinecraftClient) Close() error {
	return mc.client.Close()
}

// HelpCmd displays the help for the command \cmd|
func (mc *MinecraftClient) HelpCmd(cmd string) ([]string, error) {
	fullCmd := "help"
	if len(cmd) > 0 {
		fullCmd += " " + cmd
	}
	data, err := mc.client.Request(fullCmd)
	if err != nil {
		return []string{}, err
	}
	cmds := strings.Split(data, "/")[1:]
	return cmds, nil
}

// Help displays all of the help text.
func (mc *MinecraftClient) Help() ([]string, error) {
	return mc.HelpCmd("")
}

// UsersList lists users currently logged in to the server.
func (mc *MinecraftClient) UsersList() ([]string, error) {
	data, err := mc.client.Request("list")
	if err != nil {
		return []string{}, err
	}
	pieces := strings.SplitN(data, ":", 2)
	if len(pieces) != 2 {
		return []string{}, fmt.Errorf("invalid data: %s", data)
	}
	if len(pieces[1]) == 0 {
		return []string{}, nil
	}
	users := strings.Split(pieces[1], ",")
	for i, user := range users {
		users[i] = strings.TrimSpace(user)
	}
	return users, nil
}

// UserBan bans a user by name
func (mc *MinecraftClient) UserBan(user string) error {
	if err := validateUser(user); err != nil {
		return err
	}
	resp, err := mc.client.Request(fmt.Sprintf("ban %s", user))
	if err != nil {
		return err
	}
	return validateResponsePrefix(resp, "Banned")
}

// UserPardon pardons a user by name
func (mc *MinecraftClient) UserPardon(user string) error {
	if err := validateUser(user); err != nil {
		return err
	}
	resp, err := mc.client.Request(fmt.Sprintf("pardon %s", user))
	if err != nil {
		return err
	}
	return validateResponsePrefix(resp, "Unbanned")
}

// IPBan bans an IP address
func (mc *MinecraftClient) IPBan(ip net.IP) error {
	resp, err := mc.client.Request(fmt.Sprintf("ban-ip %s", ip.String()))
	if err != nil {
		return err
	}
	return validateResponsePrefix(resp, "Banned IP")
}

// IPPardon pardons an IP address
func (mc *MinecraftClient) IPPardon(ip net.IP) error {
	resp, err := mc.client.Request(fmt.Sprintf("pardon-ip %s", ip.String()))
	if err != nil {
		return err
	}
	return validateResponsePrefix(resp, "Unbanned IP")
}

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

type (
	RedisClient struct {
		*redis.Client
	}

	IPs struct {
		Whitelisted []string `json:"whitelist"`
	}
)

var key = "whitelist"

var RootCmd = &cobra.Command{
	Use:   "run",
	Short: "Start",
}

func Init() {
	RootCmd.PersistentFlags().String("ip", "", "User IP")
	RootCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add the user to whitelist",
		RunE: func(cmd *cobra.Command, args []string) error {
			ip, err := cmd.Flags().GetString("ip")
			if err != nil {
				return err
			}

			if net.ParseIP(ip).To4() == nil && net.ParseIP(ip).To16() == nil {
				return errors.New("Given IP should be version 4 or 6")
			}

			client := RedisClient{
				Client: redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "", // no password set
					DB:       0,  // use default DB
				}),
			}
			defer client.Close()

			client.AddIP(ip)
			fmt.Printf("Added \x1b[44m%s\x1b[0m to whitelist\r\n", ip)
			return nil
		},
	}, &cobra.Command{
		Use:   "del",
		Short: "Delete the user from whitelist",
		RunE: func(cmd *cobra.Command, args []string) error {
			ip, err := cmd.Flags().GetString("ip")
			if err != nil {
				return err
			}

			if net.ParseIP(ip).To4() == nil && net.ParseIP(ip).To16() == nil {
				return errors.New("Given IP should be version 4 or 6")
			}

			client := RedisClient{
				Client: redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "", // no password set
					DB:       0,  // use default DB
				}),
			}
			defer client.Close()

			client.DelIP(ip)
			fmt.Printf("Deleted \x1b[44m%s\x1b[0m from whitelist\r\n", ip)
			return nil
		},
	})
}

func (c *RedisClient) AddIP(ip string) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()

	users := c.GetList()

	if slices.Contains[[]string, string](users.Whitelisted, ip) {
		log.Fatal("IP already in whitelist")
	}

	list := append(users.Whitelisted, ip)
	{
		users.Whitelisted = list
	}

	serialized, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}

	c.Client.Set(ctx, key, serialized, 0)
}

func (c *RedisClient) DelIP(ip string) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()

	users := c.GetList()
	list := []string{}

	if !(slices.Contains[[]string, string](users.Whitelisted, ip)) {
		log.Fatal("IP is already not in whitelist")
	}

	for _, user := range users.Whitelisted {
		if user == ip {
			continue
		}

		list = append(list, user)
	}

	users.Whitelisted = list

	serialized, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}

	c.Client.Set(ctx, key, serialized, 0)
}

func (c *RedisClient) GetList() *IPs {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()

	users := new(IPs)
	list, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal([]byte(list), users); err != nil {
		log.Fatal(err)
	}

	if users.Whitelisted == nil {
		log.Fatal(err)
	}

	return users
}

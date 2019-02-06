package main

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/crusttech/permit/internal/api"
	"github.com/crusttech/permit/internal/rand"
	"github.com/crusttech/permit/pkg/permit"
)

type (
	validator interface {
		IsValid() bool
	}

	keeper interface {
		List(query string) ([]*permit.Permit, error)
		Get(key string) (*permit.Permit, error)
		Create(permit.Permit) error
		Revoke(key string) error
		Enable(key string) error
		Extend(key string, time *time.Time) error
		Delete(key string) error
	}
)

var domainCheck = regexp.MustCompile(`^([a-zA-Z0-9-_]+\.)*[a-zA-Z0-9][a-zA-Z0-9-_]+\.[a-zA-Z]{2,11}?$`)

func must(cmd *cobra.Command, err error) {
	if err != nil {
		cmd.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printPermit(cmd *cobra.Command, p permit.Permit) {
	cmd.Printf("Version: %d\n", p.Version)
	cmd.Printf("Key:     %s\n", p.Key)
	cmd.Printf("Domain:  %s\n", p.Domain)
	cmd.Printf("Valid:   %v\n", p.Valid)
	cmd.Printf("Expires: %s\n", p.Expires)
	if len(p.Attributes) > 0 {
		cmd.Println("---------------------------------------------")
		for name, value := range p.Attributes {
			cmd.Printf("%8d %s\n", value, name)
		}
	}
}

func commands(storage keeper) []*cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list [query]",
		Short: "List all permits",
		Long:  `pass query string to search by key`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			var q string
			if len(args) > 0 {
				q = args[0]
			}

			ll, err := storage.List(q)
			must(cmd, err)

			for _, l := range ll {
				cmd.Printf(
					"%-64s\t%-50s\t%v\t%v\n",
					l.Key,
					l.Domain,
					l.Valid,
					l.Expires,
				)
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [permit key]",
		Short: "Show single permit",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			p, err := storage.Get(args[0])
			must(cmd, err)

			printPermit(cmd, *p)
		},
	}

	createCmd := &cobra.Command{
		Use:  "create [permit domain]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var now = time.Now()
			var exp = &now

			if trial, _ := cmd.Flags().GetBool("trial"); trial {
				*exp = exp.AddDate(0, 0, 14)
			} else if inf, _ := cmd.Flags().GetBool("infinite"); !inf {
				*exp = exp.AddDate(1, 0, 0)
			} else {
				exp = nil
			}

			var key, _ = cmd.Flags().GetString("force-key")
			if key == "" {
				key = string(rand.RandBytesMaskImprSrc(permit.KeyLength))
			}

			if !domainCheck.MatchString(args[0]) {
				must(cmd, errors.New("invalid domain name format"))
			}

			p := permit.Permit{
				Version: 1,
				Expires: exp,
				Key:     key,
				Domain:  args[0],
				Valid:   true,
				Attributes: map[string]int{
					"system.enabled":                 1,
					"system.max-users":               -1,
					"system.max-organisations":       1,
					"system.max-teams":               -1,
					"messaging.enabled":              1,
					"messaging.max-users":            -1,
					"messaging.max-private-channels": -1,
					"messaging.max-public-channels":  -1,
					"messaging.max-messages":         -1,
					"compose.enabled":                1,
					"compose.max-namespaces":         -1,
					"compose.max-users":              -1,
					"compose.max-modules":            -1,
					"compose.max-charts":             -1,
					"compose.max-pages":              -1,
					"compose.max-triggers":           -1,
				},
			}

			must(cmd, storage.Create(p))

			printPermit(cmd, p)
		},
	}

	createCmd.Flags().Bool("infinite", false, "No expiration")
	createCmd.Flags().Bool("trial", false, "Trial permit")
	createCmd.Flags().String("force-key", "", "use this key instead of generated string")

	revokeCmd := &cobra.Command{
		Use:   "revoke [permit key]",
		Short: "Revokes (disables) permit",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			must(cmd, storage.Revoke(args[0]))
		},
	}

	enableCmd := &cobra.Command{
		Use:   "enable [permit key]",
		Short: "Enable permit",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			must(cmd, storage.Enable(args[0]))
		},
	}

	extendCmd := &cobra.Command{
		Use:   "extend [permit key] [duration in months]",
		Short: "Extend permit",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			months, err := strconv.Atoi(args[1])
			must(cmd, err)
			e := time.Now().AddDate(0, months, 0)
			cmd.Printf("Extending permit to %v", e)
			must(cmd, storage.Extend(args[0], &e))
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [permit key]",
		Short: "Removes permit",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			must(cmd, storage.Delete(args[0]))
		},
	}

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Removes permit",
		Run: func(cmd *cobra.Command, args []string) {
			api.Serve(storage)
		},
	}

	return []*cobra.Command{
		listCmd,
		getCmd,
		createCmd,
		revokeCmd,
		enableCmd,
		extendCmd,
		deleteCmd,
		apiCmd,
	}
}

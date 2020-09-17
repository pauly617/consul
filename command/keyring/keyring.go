package keyring

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/agent"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}

type cmd struct {
	UI    cli.Ui
	flags *flag.FlagSet
	http  *flags.HTTPFlags
	help  string

	// flags
	installKey      string
	useKey          string
	removeKey       string
	listKeys        bool
	listPrimaryKeys bool
	relay           int
	local           bool
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.installKey, "install", "",
		"Install a new encryption key. This will broadcast the new key to "+
			"all members in the cluster.")
	c.flags.StringVar(&c.useKey, "use", "",
		"Change the primary encryption key, which is used to encrypt "+
			"messages. The key must already be installed before this operation "+
			"can succeed.")
	c.flags.StringVar(&c.removeKey, "remove", "",
		"Remove the given key from the cluster. This operation may only be "+
			"performed on keys which are not currently the primary key.")
	c.flags.BoolVar(&c.listKeys, "list", false,
		"List all keys currently in use within the cluster.")
	c.flags.BoolVar(&c.listPrimaryKeys, "list-primary", false,
		"List all primary keys currently in use within the cluster.")
	c.flags.IntVar(&c.relay, "relay-factor", 0,
		"Setting this to a non-zero value will cause nodes to relay their response "+
			"to the operation through this many randomly-chosen other nodes in the "+
			"cluster. The maximum allowed value is 5.")
	c.flags.BoolVar(&c.local, "local-only", false,
		"Setting this to true will force the keyring query to only hit local servers "+
			"(no WAN traffic). This flag can only be set for list queries.")

	c.http = &flags.HTTPFlags{}
	flags.Merge(c.flags, c.http.ClientFlags())
	c.help = flags.Usage(help, c.flags)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	c.UI = &cli.PrefixedUi{
		OutputPrefix: "",
		InfoPrefix:   "==> ",
		ErrorPrefix:  "",
		Ui:           c.UI,
	}

	// Only accept a single argument
	found := c.listKeys || c.listPrimaryKeys
	for _, arg := range []string{c.installKey, c.useKey, c.removeKey} {
		if found && len(arg) > 0 {
			c.UI.Error("Only a single action is allowed")
			return 1
		}
		found = found || len(arg) > 0
	}

	// Fail fast if no actionable args were passed
	if !found {
		c.UI.Error(c.Help())
		return 1
	}

	// Validate the relay factor
	relayFactor, err := agent.ParseRelayFactor(c.relay)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error parsing relay factor: %s", err))
		return 1
	}

	// Validate local-only
	err = agent.ValidateLocalOnly(c.local, c.listKeys)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error validating local-only: %s", err))
		return 1
	}

	// All other operations will require a client connection
	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	if c.listKeys {
		c.UI.Info("Gathering installed encryption keys...")
		responses, err := client.Operator().KeyringList(&consulapi.QueryOptions{RelayFactor: relayFactor, LocalOnly: c.local})
		if err != nil {
			c.UI.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		for _, response := range responses {
			c.UI.Output(formatResponse(response, response.Keys))
		}
		return 0
	}

	if c.listPrimaryKeys {
		c.UI.Info("Gathering installed primary encryption keys...")
		responses, err := client.Operator().KeyringList(&consulapi.QueryOptions{RelayFactor: relayFactor, LocalOnly: c.local})
		if err != nil {
			c.UI.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		for _, response := range responses {
			c.UI.Output(formatResponse(response, response.PrimaryKeys))
		}
		return 0
	}

	opts := &consulapi.WriteOptions{RelayFactor: relayFactor}
	if c.installKey != "" {
		c.UI.Info("Installing new gossip encryption key...")
		err := client.Operator().KeyringInstall(c.installKey, opts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	if c.useKey != "" {
		c.UI.Info("Changing primary gossip encryption key...")
		err := client.Operator().KeyringUse(c.useKey, opts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	if c.removeKey != "" {
		c.UI.Info("Removing gossip encryption key...")
		err := client.Operator().KeyringRemove(c.removeKey, opts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	// Should never make it here
	return 0
}

func formatResponse(response *consulapi.KeyringResponse, keys map[string]int) string {
	result := []string{
		"",
		poolName(response.Datacenter, response.WAN, response.Segment) + ":",
		formatMessages(response.Messages),
		formatKeys(keys, response.NumNodes),
	}
	return strings.Replace(strings.Join(result, "\n"), "\n\n", "\n", -1)
}

func poolName(dc string, wan bool, segment string) string {
	pool := fmt.Sprintf("%s (LAN)", dc)
	if wan {
		pool = "WAN"
	}
	if segment != "" {
		segment = fmt.Sprintf(" [%s]", segment)
	}
	return fmt.Sprintf("%s%s", pool, segment)
}

func formatMessages(messages map[string]string) string {
	result := []string{}
	for from, msg := range messages {
		result = append(result, fmt.Sprintf("  ===> %s: %s", from, msg))
	}
	return strings.Join(result, "\n")
}

func formatKeys(keys map[string]int, total int) string {
	result := []string{}
	for key, num := range keys {
		result = append(result, fmt.Sprintf("  %s [%d/%d]", key, num, total))
	}
	return strings.Join(result, "\n")
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return c.help
}

const synopsis = "Manages gossip layer encryption keys"
const help = `
Usage: consul keyring [options]

  Manages encryption keys used for gossip messages. Gossip encryption is
  optional. When enabled, this command may be used to examine active encryption
  keys in the cluster, add new keys, and remove old ones. When combined, this
  functionality provides the ability to perform key rotation cluster-wide,
  without disrupting the cluster.

  All operations performed by this command can only be run against server nodes,
  and affect both the LAN and WAN keyrings in lock-step.

  All variations of the keyring command return 0 if all nodes reply and there
  are no errors. If any node fails to reply or reports failure, the exit code
  will be 1.
`

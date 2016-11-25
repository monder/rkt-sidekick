package modes

import (
	"flag"
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/mitchellh/cli"
	"github.com/monder/rkt-sidekick/lib"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type command struct {
	etcdAddress string
	cidr        string
	format      string
	expireDir   string
	interval    time.Duration
}

var EtcdCommand = func() (cli.Command, error) {
	return &command{}, nil
}

func (c command) Help() string {
	return `
Usage: rkt-sidekick [options] [KEY_IN_ETCD]

Options:
    -cidr              cidr to match the ip (default: "0.0.0.0/0")
    -etcd-endpoint     an etcd address in the cluster (default: "http://172.16.28.1:2379")
    -format            format of the etcd key value. '$ip' will be replace by container's ip address (default: "$ip")
    -expire-dir        set expiration TTLs for all items under that directory, not only the leaf node
    -interval          refresh interval (default: "1m")
`
}

func (c command) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("etcd", flag.ContinueOnError)
	cmdFlags.Usage = func() { fmt.Fprintf(os.Stderr, "%s", c.Help()) }

	cmdFlags.StringVar(&c.etcdAddress, "etcd-endpoint", "http://172.16.28.1:2379", "")
	cmdFlags.StringVar(&c.cidr, "cidr", "0.0.0.0/0", "")
	cmdFlags.StringVar(&c.format, "format", "$ip", "")
	cmdFlags.StringVar(&c.expireDir, "expire-dir", "", "")
	cmdFlags.DurationVar(&c.interval, "interval", time.Minute, "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	key := ""

	// Check for arg validation
	args = cmdFlags.Args()
	switch len(args) {
	case 0:
		key = ""
	case 1:
		key = args[0]
	default:
		fmt.Fprintf(os.Stderr, "Too many arguments (expected 1, got %d)\n", len(args))
		return 1
	}
	if key == "" {
		fmt.Fprintf(os.Stderr, "Error! Missing KEY argument\n")
		return 1
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)

	etcd, err := getEtcd(strings.Split(c.etcdAddress, ","))
	if err != nil {
		log.Fatal(err)
	}
	ip, err := lib.GetIPAddress(c.cidr)
	if err != nil {
		log.Fatal(err)
	}

	value := strings.Replace(c.format, "$ip", ip, -1)

	_, err = etcd.Set(context.Background(), key, value, &client.SetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(c.interval)

	for {
		path := key
		keepRoot := c.expireDir
		if keepRoot == "" {
			keepRoot = path[0:strings.LastIndex(path, "/")]
		}
		if !strings.HasSuffix(keepRoot, "/") {
			keepRoot += "/"
		}
		currentPath := "/"
		if strings.HasPrefix(path, keepRoot) {
			path = path[len(keepRoot):]
			currentPath = keepRoot
		} else {
			keepRoot = currentPath
			path = path[len(currentPath):]
		}

		for _, s := range strings.SplitAfter(path, "/") {
			currentPath += s
			fmt.Printf("Setting ttl for %s (dir: %s)\n", currentPath, strings.HasSuffix(currentPath, "/"))
			_, err = etcd.Set(context.Background(), currentPath, "", &client.SetOptions{
				Refresh:   true,
				PrevExist: client.PrevExist,
				TTL:       2 * c.interval,
				Dir:       strings.HasSuffix(currentPath, "/"),
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		select {
		case <-sigc:
			etcd.Delete(context.Background(), key, nil)
			//TODO clear empty dirs
			os.Exit(0)
		case <-ticker.C:
		}
	}
	return 0
}

func getEtcd(endpoints []string) (client.KeysAPI, error) {
	cfg := client.Config{
		Endpoints: endpoints,
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	kapi := client.NewKeysAPI(c)
	return kapi, nil
}

func (c command) Synopsis() string {
	return "Updates a key in etcd with container IP address"
}

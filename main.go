package main

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var flags struct {
	etcdAddress string
	cidr        string
	format      string
	expireDir   string
	interval    time.Duration
}

func init() {
	pflag.StringVarP(&flags.etcdAddress, "etcd-endpoint", "e", "http://172.16.28.1:2379", "an etcd address in the cluster")
	pflag.StringVar(&flags.cidr, "cidr", "0.0.0.0/0", "cidr to match the ip")
	pflag.StringVarP(&flags.format, "format", "f", "$ip", "format of the etcd key value. '$ip' will be replace by container's ip address")
	pflag.StringVar(&flags.expireDir, "expireDir", "", "set expiration TTLs for all items under that directory, not only the leaf node")
	pflag.DurationVarP(&flags.interval, "interval", "i", time.Minute, "refresh interval")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s /key/in/etcd\n\nFlags:\n", os.Args[0])
		pflag.PrintDefaults()
	}
}

func main() {
	pflag.Parse()
	if pflag.Arg(0) == "" {
		pflag.Usage()
		os.Exit(2)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)

	etcd, err := getEtcd([]string{flags.etcdAddress})
	if err != nil {
		log.Fatal(err)
	}
	ip, err := getIPAddress(flags.cidr)
	if err != nil {
		log.Fatal(err)
	}

	value := strings.Replace(flags.format, "$ip", ip, -1)

	_, err = etcd.Set(context.Background(), pflag.Arg(0), value, &client.SetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(flags.interval)

	for {
		path := pflag.Arg(0)
		keepRoot := flags.expireDir
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
				TTL:       2 * flags.interval,
				Dir:       strings.HasSuffix(currentPath, "/"),
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		select {
		case <-sigc:
			etcd.Delete(context.Background(), pflag.Arg(0), nil)
			//TODO clear empty dirs
			os.Exit(0)
		case <-ticker.C:
		}
	}
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

func getIPAddress(cidr string) (string, error) {
	_, requirednet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if requirednet.Contains(ipnet.IP) {
				return ipnet.IP.String(), nil
			}
		}

	}
	return "", fmt.Errorf("No IP address matching %s found.", cidr)
}

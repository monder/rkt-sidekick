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
	"syscall"
	"time"
)

var flags struct {
	etcdAddress string
	cidr        string
	interval    time.Duration
}

func init() {
	pflag.StringVarP(&flags.etcdAddress, "etcd-endpoint", "e", "http://172.16.28.1:2379", "an etcd address in the cluster")
	pflag.StringVar(&flags.cidr, "cidr", "0.0.0.0/0", "cidr to match the ip")
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

	_, err = etcd.Set(context.Background(), pflag.Arg(0), ip, &client.SetOptions{TTL: 2 * flags.interval})
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(flags.interval)

	for {
		_, err = etcd.Set(context.Background(), pflag.Arg(0), "", &client.SetOptions{Refresh: true, TTL: 2 * flags.interval})
		if err != nil {
			log.Fatal(err)
		}
		select {
		case <-sigc:
			etcd.Delete(context.Background(), pflag.Arg(0), nil)
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

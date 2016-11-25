package modes

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/mitchellh/cli"
	"github.com/monder/rkt-sidekick/lib"
	"log"
	"os"
)

type route53Command struct {
	cidr string
}

var Route53Command = func() (cli.Command, error) {
	return &route53Command{}, nil
}

func (c route53Command) Help() string {
	return `
Usage: rkt-sidekick route53 [options] [ZONE_ID] [HOSTNAME]

Options:
    -cidr              cidr to match the ip (default: "0.0.0.0/0")
`
}

func (c route53Command) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("route53", flag.ContinueOnError)
	cmdFlags.Usage = func() { fmt.Fprintf(os.Stderr, "%s", c.Help()) }
	cmdFlags.StringVar(&c.cidr, "cidr", "0.0.0.0/0", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	hostname := ""
	zoneId := ""

	// Check for arg validation
	args = cmdFlags.Args()
	switch len(args) {
	case 2:
		zoneId = args[0]
		hostname = args[1]
	default:
		fmt.Fprintf(os.Stderr, "Incorrect number of arguments (expected 2, got %d)\n", len(args))
		return 1
	}
	if hostname == "" {
		fmt.Fprintf(os.Stderr, "Error! Missing HOSTNAME argument\n")
		return 1
	}
	ip, err := lib.GetIPAddress(c.cidr)
	if err != nil {
		log.Fatal(err)
		return 2
	}

	r := route53.New(session.New())
	_, err = r.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(hostname),
						Type: aws.String(route53.RRTypeA),
						TTL:  aws.Int64(1),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip),
							},
						},
					},
				},
			},
		},
		HostedZoneId: aws.String(zoneId),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	return 0
}

func (c route53Command) Synopsis() string {
	return "Updates the route53 endpoint with container IP address"
}

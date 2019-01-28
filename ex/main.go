package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/route53"
)

// type Configuration struct {
// 	Port              int
// 	Static_Variable   string
// 	Port2             int
// 	Static_Variable2  string
// 	Connection_String string
// }

type Configuration struct {
	Max             int64
	Min             int64
	DesiredCapacity int64
	Auto            string
	DNSName         string
	Name            string
	HostZone        string
}

func init() {
	// configuration := Configuration{}

	// flag.Int64Var(&configuration.Max, "max", 0, "maximum size for an autoscaling group")
	// flag.Int64Var(&configuration.Min, "min", 0, "minuim size for an autoscaling group")
	// flag.Int64Var(&configuration.DesiredCapacity, "desired", 0, "desired capacity for an autoscaling group")
	// flag.StringVar(&configuration.Auto, "auto", "", "autoscaling group name")
	// flag.StringVar(&configuration.DNSName, "dnsname", "", "DNS record set")
	// flag.StringVar(&configuration.Name, "name", "", "replaced DNS record set")
	// flag.StringVar(&configuration.HostZone, "zone", "", "target zone that has the record sets")
}
func main() {

	configuration := Configuration{}

	flag.Int64Var(&configuration.Max, "max", 0, "maximum size for an autoscaling group")
	flag.Int64Var(&configuration.Min, "min", 0, "minuim size for an autoscaling group")
	flag.Int64Var(&configuration.DesiredCapacity, "desired", 0, "desired capacity for an autoscaling group")
	flag.StringVar(&configuration.Auto, "auto", "", "autoscaling group name")
	flag.StringVar(&configuration.DNSName, "dnsname", "", "DNS record set")
	flag.StringVar(&configuration.Name, "name", "", "replaced DNS record set")
	flag.StringVar(&configuration.HostZone, "zone", "", "target zone that has the record sets")

	flag.Parse()

	st := []string{"development.json", "production.json"}

	for v := range st {
		filename := "./config." + st[v]

		file, err := os.Open(filename)
		if err != nil {
			log.Fatalln(err)
		}
		configuration := Configuration{}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&configuration)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(configuration.Max)
		fmt.Println(configuration.DNSName)

		updateAutoScaling(configuration.Max, configuration.Min, configuration.DesiredCapacity, configuration.Name, configuration.Auto)
		updateRecordSets(configuration.DNSName, configuration.HostZone, configuration.Name)
	}

}
func updateAutoScaling(max, min, desired int64, name, auto string) {

	svc := autoscaling.New(session.New())
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(auto),
		MaxSize:              &max,
		MinSize:              &min,
		DesiredCapacity:      &desired,
	}

	result, err := svc.UpdateAutoScalingGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeScalingActivityInProgressFault:
				fmt.Println(autoscaling.ErrCodeScalingActivityInProgressFault, aerr.Error())
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
			case autoscaling.ErrCodeLimitExceededFault:
				fmt.Println(autoscaling.ErrCodeLimitExceededFault, aerr.Error())

			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println(result)
}

func updateRecordSets(dNSName, hostZone, name string) {
	svc := route53.New(session.New())
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						AliasTarget: &route53.AliasTarget{
							DNSName:              aws.String(dNSName),
							EvaluateTargetHealth: aws.Bool(false),
							HostedZoneId:         aws.String(hostZone),
						},
						Name: aws.String(name),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("Switch dns name for testing purposes"),
		},
		HostedZoneId: aws.String(hostZone),
	}

	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				fmt.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				fmt.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				fmt.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				fmt.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				fmt.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println(result)
}

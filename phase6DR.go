package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Syncbak-Git/jsconfig"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	maxSize         int64
	minSize         int64
	desiredCapacity int64
	auto            string
	desc            string
	dNSName         string
	name            string
	hostZone        string
	configFiles     string
	// version         bool
	// help            bool
)

//const versionStr = "0.0.1"

func init() {

	// flag.StringVar(&configFiles, "c", "./config/default.json", "Comma-seperated list of json config files")
	// flag.BoolVar(&version, "v", false, "print the version and exit the program")
	// flag.BoolVar(&version, "version", false, "print the version adn exit the program")
	// flag.BoolVar(&help, "h", false, "prints this helpful text")

	// flag.Int64Var(&maxSize, "max", 0, "maximum size for an autoscaling group")
	// flag.Int64Var(&minSize, "min", 0, "minuim size for an autoscaling group")
	// flag.Int64Var(&desiredCapacity, "desired", 0, "desired capacity for an autoscaling group")
	// flag.StringVar(&auto, "auto", "", "autoscaling group name")
	// flag.StringVar(&desc, "desc", "", "describe an autoscling group")
	// flag.StringVar(&dNSName, "dnsname", "", "DNS record set")
	// flag.StringVar(&name, "name", "", "replaced DNS record set")
	// flag.StringVar(&hostZone, "zone", "", "target zone that has the record sets")
}

func main() {

	// if version {
	// 	fmt.Printf("%s\n", versionStr)
	// 	os.Exit(0)
	// }
	//log.Info("config files %s", configFiles)
	err := jsconfig.InitFromFiles("./config/default.json")
	if err != nil {
		log.Fatalln("Could not read config setting from ", err)
	}

	// err = jsconfig.InitFromFiles("./config/config.json")
	// if err != nil {
	// 	log.Fatalln("Could not read config setting from ", err)
	// }

	flag.Parse()

	maxSize = int64(jsconfig.S.FindNumber("max"))
	minSize = int64(jsconfig.S.FindNumber("min"))
	desiredCapacity = int64(jsconfig.S.FindNumber("desired"))
	auto = jsconfig.S.FindString("auto")
	dNSName = jsconfig.S.FindString("dnsname")
	name = jsconfig.S.FindString("name")
	hostZone = jsconfig.S.FindString("hostZone")

	//describeRateLimit()
	//describeAutoScaling()
	updateAutoScaling(minSize, maxSize, desiredCapacity, name, auto)
	updateRecordSets(dNSName, hostZone, name)
}

func describeRateLimit() {
	svc := autoscaling.New(session.New())
	input := &autoscaling.DescribeAccountLimitsInput{}

	result, err := svc.DescribeAccountLimits(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
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

func describeAutoScaling() {
	svc := autoscaling.New(session.New())
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(desc),
		},
	}
	result, err := svc.DescribeAutoScalingGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeInvalidNextToken:
				fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
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

func updateAutoScaling(max, min, desired int64, name, auto string) {

	svc := autoscaling.New(session.New())
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(auto),
		MaxSize:              &maxSize,
		MinSize:              &minSize,
		DesiredCapacity:      &desiredCapacity,
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

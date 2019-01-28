package main

import (
	"flag"
	"fmt"

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
)

func init() {

	flag.Int64Var(&maxSize, "max", 20, "maximum size for an autoscaling group")
	flag.Int64Var(&minSize, "min", 2, "minuim size for an autoscaling group")
	flag.Int64Var(&desiredCapacity, "desired", 2, "desired capacity for an autoscaling group")
	flag.StringVar(&auto, "auto", "phase6-VPC-qa-v010", "autoscaling group name")
	flag.StringVar(&dNSName, "dnsname", "phase6test.aws.syncbak.com", "DNS record set")
	flag.StringVar(&name, "name", "phase6dr.aws.syncbak.com", "replaced DNS record set")
	flag.StringVar(&hostZone, "zone", "Z219GR296HPKS6", "target zone that has the record sets")

}

func main() {

	flag.Parse()

	updateAutoScaling(minSize, maxSize, desiredCapacity, name, auto)
	//updateRecordSets(dNSName, hostZone, name)
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

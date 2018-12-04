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

var MaxSize int64
var MinSize int64
var DesiredCapacity int64
var Auto string
var Desc string
var DNSName string
var Name string
var HosteZone string

func main() {

	flag.Int64Var(&MaxSize, "max", 1, "maximum size for an autoscaling group")
	flag.Int64Var(&MinSize, "min", 1, "minuim size for an autoscaling group")
	flag.Int64Var(&DesiredCapacity, "desired", 1, "desired capacity for an autoscaling group")
	flag.StringVar(&Auto, "auto", "phase6-VPC-cdnadapter-EnvFile-d0qa-v079", "autoscaling group name")
	flag.StringVar(&Desc, "desc", "phase6-VPC-cdnadapter-EnvFile-d0qa-v079", "describe an autoscling group")
	flag.StringVar(&DNSName, "dnsname", "phase6test.aws.syncbak.com", "replaced record set")
	flag.StringVar(&Name, "name", "phase6dr.aws.syncbak.com", "replaced with record set")
	flag.StringVar(&HosteZone, "zone", "Z219GR296HPKS6", "target zone that has the record sets")

	flag.Parse()

	describeRateLimit()
	describeAutoScaling()
	updateAutoScaling()
	updateRecordSets()
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
			aws.String(Desc),
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

func updateAutoScaling() {

	svc := autoscaling.New(session.New())
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(Auto),
		MaxSize:              &MaxSize,
		MinSize:              &MinSize,
		DesiredCapacity:      &DesiredCapacity,
	}

	result, err := svc.UpdateAutoScalingGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeScalingActivityInProgressFault:
				fmt.Println(autoscaling.ErrCodeScalingActivityInProgressFault, aerr.Error())
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
			//case autoscaling.ErrCodeServiceLinkedRoleFailure:
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

func updateRecordSets() {
	svc := route53.New(session.New())
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						AliasTarget: &route53.AliasTarget{
							DNSName:              aws.String(DNSName),
							EvaluateTargetHealth: aws.Bool(false),
							HostedZoneId:         aws.String(HosteZone),
						},
						Name: aws.String(Name),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("Switch dns name for testing purposes"),
		},
		HostedZoneId: aws.String(HosteZone),
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

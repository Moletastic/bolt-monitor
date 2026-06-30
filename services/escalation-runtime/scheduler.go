package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type scheduleClient interface {
	ScheduleNextStep(context.Context, scheduledInvocationEvent, time.Time) error
}

type eventBridgeAPI interface {
	PutRule(context.Context, *eventbridge.PutRuleInput, ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error)
	PutTargets(context.Context, *eventbridge.PutTargetsInput, ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error)
}

type lambdaAPI interface {
	AddPermission(context.Context, *awslambda.AddPermissionInput, ...func(*awslambda.Options)) (*awslambda.AddPermissionOutput, error)
}

type stsAPI interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type cloudWatchScheduler struct {
	eventsClient eventBridgeAPI
	lambdaClient lambdaAPI
	stsClient    stsAPI
	region       string
	functionName string
}

type scheduledInvocationEvent struct {
	IncidentID string `json:"incidentId"`
	Step       int    `json:"step"`
}

func newCloudWatchScheduler(eventsClient eventBridgeAPI, lambdaClient lambdaAPI, stsClient stsAPI, region, functionName string) *cloudWatchScheduler {
	return &cloudWatchScheduler{eventsClient: eventsClient, lambdaClient: lambdaClient, stsClient: stsClient, region: region, functionName: functionName}
}

func (s *cloudWatchScheduler) ScheduleNextStep(ctx context.Context, event scheduledInvocationEvent, when time.Time) error {
	if strings.TrimSpace(s.region) == "" || strings.TrimSpace(s.functionName) == "" {
		return fmt.Errorf("scheduler requires AWS region and function name")
	}
	accountID, err := s.accountID(ctx)
	if err != nil {
		return err
	}
	ruleName := scheduledRuleName(event.IncidentID, event.Step)
	ruleOut, err := s.eventsClient.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:               aws.String(ruleName),
		ScheduleExpression: aws.String(scheduleExpressionForTime(when)),
		State:              eventbridgetypes.RuleStateEnabled,
	})
	if err != nil {
		return err
	}
	functionArn := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", s.region, accountID, s.functionName)
	statementID := "allow-events-" + strings.ToLower(strings.ReplaceAll(ruleName, "_", "-"))
	_, err = s.lambdaClient.AddPermission(ctx, &awslambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),
		FunctionName: aws.String(s.functionName),
		Principal:    aws.String("events.amazonaws.com"),
		StatementId:  aws.String(statementID),
		SourceArn:    ruleOut.RuleArn,
	})
	if err != nil && !strings.Contains(err.Error(), "ResourceConflictException") {
		return err
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = s.eventsClient.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []eventbridgetypes.Target{{
			Arn:   aws.String(functionArn),
			Id:    aws.String("1"),
			Input: aws.String(string(body)),
		}},
	})
	return err
}

func (s *cloudWatchScheduler) accountID(ctx context.Context) (string, error) {
	out, err := s.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.Account), nil
}

func scheduledRuleName(incidentID string, step int) string {
	return "esc-" + strings.ToLower(strings.TrimSpace(incidentID)) + "-step-" + fmt.Sprintf("%d", step)
}

func scheduleExpressionForTime(when time.Time) string {
	utc := when.UTC()
	return fmt.Sprintf("cron(%d %d %d %d ? %d)", utc.Minute(), utc.Hour(), utc.Day(), int(utc.Month()), utc.Year())
}

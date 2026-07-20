package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type ruleSummary struct {
	Name string
	Arn  string
}

type statementSummary struct {
	StatementID string
	Action      string
	Principal   string
}

type cleanupReport struct {
	DryRun           bool               `json:"dryRun"`
	Rules            []ruleSummary      `json:"rules"`
	Statements       []statementSummary `json:"statements"`
	SkippedRule      []string           `json:"skippedRule,omitempty"`
	SkippedStatement []string           `json:"skippedStatement,omitempty"`
	Errors           []string           `json:"errors,omitempty"`
}

func runLegacyCleanup(ctx context.Context, apply bool) error {
	cfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	eb := eventbridge.NewFromConfig(cfg)
	lambdaClient := lambda.NewFromConfig(cfg)

	report := cleanupReport{DryRun: !apply}

	rules, skipped, err := listLegacyRules(ctx, eb)
	if err != nil {
		return err
	}
	report.Rules = rules
	report.SkippedRule = skipped

	statements, skippedStatements, err := listLegacyStatements(ctx, lambdaClient)
	if err != nil {
		return err
	}
	report.Statements = statements
	report.SkippedStatement = skippedStatements

	if apply {
		for _, rule := range rules {
			if _, err := eb.DeleteRule(ctx, &eventbridge.DeleteRuleInput{Name: aws.String(rule.Name)}); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("rule %s: %v", rule.Name, err))
			}
		}
		for _, stmt := range statements {
			if _, err := lambdaClient.RemovePermission(ctx, &lambda.RemovePermissionInput{
				FunctionName: aws.String(currentFunctionName()),
				StatementId:  aws.String(stmt.StatementID),
			}); err != nil {
				var notFound *lambdatypes.ResourceNotFoundException
				if errors.As(err, &notFound) {
					continue
				}
				report.Errors = append(report.Errors, fmt.Sprintf("statement %s: %v", stmt.StatementID, err))
			}
		}
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if _, err := os.Stdout.Write(encoded); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func listLegacyRules(ctx context.Context, eb *eventbridge.Client) ([]ruleSummary, []string, error) {
	rulesOutput, err := eb.ListRules(ctx, &eventbridge.ListRulesInput{})
	if err != nil {
		return nil, nil, fmt.Errorf("list rules: %w", err)
	}
	var matches []ruleSummary
	var skipped []string
	for _, rule := range rulesOutput.Rules {
		name := aws.ToString(rule.Name)
		if !strings.HasPrefix(name, "esc-") || !strings.Contains(name, "-step-") {
			continue
		}
		matches = append(matches, ruleSummary{Name: name, Arn: aws.ToString(rule.Arn)})
	}
	return matches, skipped, nil
}

func listLegacyStatements(ctx context.Context, client *lambda.Client) ([]statementSummary, []string, error) {
	policyOutput, err := client.GetPolicy(ctx, &lambda.GetPolicyInput{
		FunctionName: aws.String(currentFunctionName()),
	})
	if err != nil {
		var notFound *lambdatypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("get lambda policy: %w", err)
	}
	policy := aws.ToString(policyOutput.Policy)
	if policy == "" {
		return nil, nil, nil
	}
	parsed, err := parseLambdaPolicy(policy)
	if err != nil {
		return nil, nil, err
	}
	var matches []statementSummary
	var skipped []string
	for _, stmt := range parsed {
		if !strings.HasPrefix(stmt.StatementID, "allow-events-") {
			continue
		}
		if !strings.Contains(stmt.Action, "lambda:InvokeFunction") {
			continue
		}
		if stmt.Principal != "events.amazonaws.com" {
			continue
		}
		matches = append(matches, stmt)
	}
	return matches, skipped, nil
}

func currentFunctionName() string {
	if v := os.Getenv("LEGACY_LAMBDA_FUNCTION"); v != "" {
		return v
	}
	if v := os.Getenv("AWS_LAMBDA_FUNCTION_NAME"); v != "" {
		return v
	}
	return "bolt-monitor-staging-escalation-runtime"
}

func parseLambdaPolicy(raw string) ([]statementSummary, error) {
	var doc struct {
		Statement []struct {
			Sid       string `json:"Sid"`
			Action    string `json:"Action"`
			Principal string `json:"Principal"`
		} `json:"Statement"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("parse lambda policy: %w", err)
	}
	out := make([]statementSummary, 0, len(doc.Statement))
	for _, stmt := range doc.Statement {
		out = append(out, statementSummary{
			StatementID: stmt.Sid,
			Action:      stmt.Action,
			Principal:   stmt.Principal,
		})
	}
	return out, nil
}

func matchLegacyStatements(statements []statementSummary) []statementSummary {
	var matches []statementSummary
	for _, stmt := range statements {
		if !strings.HasPrefix(stmt.StatementID, "allow-events-") {
			continue
		}
		if !strings.Contains(stmt.Action, "lambda:InvokeFunction") {
			continue
		}
		if stmt.Principal != "events.amazonaws.com" {
			continue
		}
		matches = append(matches, stmt)
	}
	return matches
}

func main() {
	apply := flag.Bool("apply", false, "actually remove matching legacy resources; default is dry-run")
	flag.Parse()

	if err := runLegacyCleanup(context.Background(), *apply); err != nil {
		log.Fatalf("legacy-eventbridge-cleanup: %v", err)
	}
}

var _ eventbridgetypes.Rule

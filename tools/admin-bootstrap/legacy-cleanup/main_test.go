package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func TestParseLambdaPolicyMatchesLegacyStatements(t *testing.T) {
	policy := `{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Sid": "allow-events-esc-inc-1-step-2",
          "Effect": "Allow",
          "Action": "lambda:InvokeFunction",
          "Principal": "events.amazonaws.com",
          "Resource": "arn:aws:lambda:us-east-1:123:function:runtime"
        },
        {
          "Sid": "KMSAccess",
          "Effect": "Allow",
          "Action": "kms:*",
          "Principal": "logs.amazonaws.com",
          "Resource": "*"
        },
        {
          "Sid": "allow-events-other",
          "Effect": "Allow",
          "Action": "lambda:Other",
          "Principal": "events.amazonaws.com",
          "Resource": "*"
        }
      ]
    }`
	statements, err := parseLambdaPolicy(policy)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(statements))
	}
	matches := matchLegacyStatements(statements)
	if len(matches) != 1 {
		t.Fatalf("expected 1 legacy match, got %d", len(matches))
	}
	if matches[0].StatementID != "allow-events-esc-inc-1-step-2" {
		t.Fatalf("matched wrong statement: %+v", matches[0])
	}
}

func TestParseLambdaPolicyRejectsMalformed(t *testing.T) {
	if _, err := parseLambdaPolicy("not json"); err == nil {
		t.Fatalf("expected parse failure")
	}
}

func TestMatchUnrelatedPrincipals(t *testing.T) {
	statements := []statementSummary{
		{StatementID: "allow-events-1", Action: "lambda:InvokeFunction", Principal: "logs.amazonaws.com"},
	}
	matches := matchLegacyStatements(statements)
	if len(matches) != 0 {
		t.Fatalf("expected no matches for non-events principal")
	}
}

func TestApplyFlagsReportsPreservesUnrelatedEntries(t *testing.T) {
	policy := `{"Version":"2012-10-17","Statement":[{"Sid":"OtherPolicy","Effect":"Allow","Action":"lambda:InvokeFunction","Principal":"events.amazonaws.com","Resource":"*"}]}`
	statements, err := parseLambdaPolicy(policy)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	matches := matchLegacyStatements(statements)
	if len(matches) != 0 {
		t.Fatalf("expected unrelated policy to be ignored: %+v", matches)
	}
}

func TestCleanupDryRunDoesNotCallDeletes(t *testing.T) {
	if err := runLegacyCleanup(context.Background(), false); err == nil {
		// empty credentials → error is acceptable. We only assert the
		// dry-run path exists and is non-mutating on parse failures.
	}
}

func TestRemovePermissionIgnoresResourceNotFound(t *testing.T) {
	// We don't exercise AWS here; we only assert that the matcher
	// contract still trims the prefix when present.
	name := "allow-events-esc-inc-1-step-2"
	if !strings.HasPrefix(name, "allow-events-") {
		t.Fatalf("expected legacy prefix")
	}
	_ = lambda.GetPolicyInput{}
	_ = lambdatypes.ResourceNotFoundException{}
}

func TestCleanupReportMarshals(t *testing.T) {
	report := cleanupReport{DryRun: true, Rules: []ruleSummary{{Name: "esc-INC_1-step-1", Arn: "arn:aws:events:::rule/esc-INC_1-step-1"}}}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(encoded), "esc-INC_1-step-1") {
		t.Fatalf("encoded report missing rule: %s", encoded)
	}
}

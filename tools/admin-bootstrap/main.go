package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const unknownStage = "unknown"

var stagePattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,128}$`)

type bootstrapOutcome struct {
	Timestamp          string               `json:"timestamp"`
	Event              auth.SecurityEvent   `json:"event"`
	Outcome            string               `json:"outcome"`
	Stage              string               `json:"stage"`
	ActingAWSPrincipal string               `json:"actingAwsPrincipal,omitempty"`
	TargetSubject      string               `json:"targetSubject,omitempty"`
	DesiredAuthority   bootstrapAuthority   `json:"desiredAuthority"`
	Correlation        bootstrapCorrelation `json:"correlation"`
}

type bootstrapAuthority struct {
	TenantID string `json:"tenantId"`
	Status   string `json:"status"`
	Role     string `json:"role"`
}

type bootstrapCorrelation struct {
	ID string `json:"id"`
}

func main() {
	email := flag.String("email", "", "operator email")
	userPoolID := flag.String("user-pool-id", "", "Cognito user pool ID")
	authTable := flag.String("auth-table", "", "AuthTable name")
	stage := flag.String("stage", os.Getenv("SST_STAGE"), "deployment stage")
	flag.Parse()
	if *email == "" || *userPoolID == "" || *authTable == "" {
		log.Fatal("email, user-pool-id, and auth-table are required")
	}

	ctx := context.Background()
	correlationID, err := newCorrelationID()
	if err != nil {
		log.Fatalf("generate bootstrap correlation id: %v", err)
	}
	resolvedStage, stageErr := normalizeStage(*stage)
	actingPrincipal := lookupActingAWSPrincipal(ctx)
	if stageErr != nil {
		emitOutcomeOrFatal(os.Stdout, newBootstrapOutcome(unknownStage, actingPrincipal, "", correlationID, stageErr))
		log.Fatal("bootstrap administrator: invalid stage")
	}
	cognito, err := sharedaws.NewCognitoIdentityProviderAPI(ctx)
	if err != nil {
		emitOutcomeOrFatal(os.Stdout, newBootstrapOutcome(resolvedStage, actingPrincipal, "", correlationID, err))
		log.Fatal("bootstrap administrator failed while creating Cognito client")
	}
	dynamo, err := sharedaws.NewDynamoDBAPI(ctx)
	if err != nil {
		emitOutcomeOrFatal(os.Stdout, newBootstrapOutcome(resolvedStage, actingPrincipal, "", correlationID, err))
		log.Fatal("bootstrap administrator failed while creating DynamoDB client")
	}
	bootstrap := bootstrapper{cognito: cognito, dynamo: dynamo, userPoolID: *userPoolID, authTable: *authTable, now: time.Now, membershipID: newMembershipID}
	subject, err := bootstrap.bootstrapWithEvents(ctx, *email, func(event auth.SecurityEvent, subject auth.Subject) {
		emitOutcomeOrFatal(os.Stdout, newSecurityOutcome(event, resolvedStage, actingPrincipal, subject, correlationID, nil))
	})
	if err != nil {
		emitOutcomeOrFatal(os.Stdout, newBootstrapOutcome(resolvedStage, actingPrincipal, subject, correlationID, err))
		log.Fatal("bootstrap administrator failed")
	}
	emitOutcomeOrFatal(os.Stdout, newBootstrapOutcome(resolvedStage, actingPrincipal, subject, correlationID, nil))
}

func newBootstrapOutcome(stage, actingPrincipal string, subject auth.Subject, correlationID string, err error) bootstrapOutcome {
	return newSecurityOutcome(auth.EventBootstrapReconciled, stage, actingPrincipal, subject, correlationID, err)
}

func newSecurityOutcome(event auth.SecurityEvent, stage, actingPrincipal string, subject auth.Subject, correlationID string, err error) bootstrapOutcome {
	outcome := "success"
	if err != nil {
		outcome = "failure"
	}
	return bootstrapOutcome{
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		Event:              event,
		Outcome:            outcome,
		Stage:              stage,
		ActingAWSPrincipal: actingPrincipal,
		TargetSubject:      string(subject),
		DesiredAuthority: bootstrapAuthority{
			TenantID: string(auth.DefaultTenantID),
			Status:   string(auth.MembershipStatusActive),
			Role:     string(auth.RoleAdmin),
		},
		Correlation: bootstrapCorrelation{ID: correlationID},
	}
}

func emitOutcomeOrFatal(writer io.Writer, outcome bootstrapOutcome) {
	if err := json.NewEncoder(writer).Encode(outcome); err != nil {
		log.Fatalf("emit bootstrap outcome: %v", err)
	}
}

func normalizeStage(stage string) (string, error) {
	stage = strings.TrimSpace(stage)
	if !stagePattern.MatchString(stage) {
		return "", fmt.Errorf("invalid stage")
	}
	return stage, nil
}

func newCorrelationID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func lookupActingAWSPrincipal(ctx context.Context) string {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ""
	}
	out, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil || out.Arn == nil {
		return ""
	}
	return *out.Arn
}

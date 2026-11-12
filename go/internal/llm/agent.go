// Package llm is the only package that imports the Anthropic SDK.
package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

// DefaultModel is used when ANTHROPIC_MODEL is not set.
const DefaultModel = "claude-opus-5-5"

// AgentSpec is the Managed Agent the reviewer runs on. Change it here and
// re-run `make agent-setup`; each change creates a new agent version.
type AgentSpec struct {
	Name   string
	Model  string
	System string
}

var ReviewerAgent = AgentSpec{
	Name:   "Document Reviewer",
	Model:  DefaultModel,
	System: "You are a helpful assistant.",
}

const environmentName = "document-review"

type Client struct {
	api anthropic.Client
}

// NewClient resolves credentials the standard SDK way (ANTHROPIC_API_KEY,
// or an `ant auth login` profile).
func NewClient() *Client {
	return &Client{api: anthropic.NewClient()}
}

// EnsureEnvironment returns environmentID if it exists, or creates a new
// cloud environment with limited networking.
func (c *Client) EnsureEnvironment(ctx context.Context, environmentID string) (string, bool, error) {
	if environmentID != "" {
		env, err := c.api.Beta.Environments.Get(ctx, environmentID, anthropic.BetaEnvironmentGetParams{})
		if err != nil {
			return "", false, fmt.Errorf("retrieving environment %s: %w", environmentID, err)
		}
		return env.ID, false, nil
	}
	env, err := c.api.Beta.Environments.New(ctx, anthropic.BetaEnvironmentNewParams{
		Name: environmentName,
		Config: anthropic.BetaEnvironmentNewParamsConfigUnion{
			OfCloud: &anthropic.BetaCloudConfigParams{
				Networking: anthropic.BetaCloudConfigParamsNetworkingUnion{
					OfLimited: &anthropic.BetaLimitedNetworkParams{},
				},
			},
		},
	})
	if err != nil {
		return "", false, fmt.Errorf("creating environment: %w", err)
	}
	return env.ID, true, nil
}

// EnsureAgent updates agentID to spec, or creates the agent if agentID is
// empty. It returns the agent ID and its current version.
func (c *Client) EnsureAgent(ctx context.Context, agentID string, spec AgentSpec) (string, int64, error) {
	if agentID != "" {
		current, err := c.api.Beta.Agents.Get(ctx, agentID, anthropic.BetaAgentGetParams{})
		if err != nil {
			return "", 0, fmt.Errorf("retrieving agent %s: %w", agentID, err)
		}
		updated, err := c.api.Beta.Agents.Update(ctx, agentID, anthropic.BetaAgentUpdateParams{
			Version: anthropic.Int(current.Version),
			Name:    anthropic.String(spec.Name),
			Model:   anthropic.BetaManagedAgentsModelConfigParams{ID: spec.Model},
			System:  anthropic.String(spec.System),
			Tools:   []anthropic.BetaAgentUpdateParamsToolUnion{{OfAgentToolset20260401: toolset()}},
		})
		if err != nil {
			return "", 0, fmt.Errorf("updating agent %s: %w", agentID, err)
		}
		return updated.ID, updated.Version, nil
	}
	created, err := c.api.Beta.Agents.New(ctx, anthropic.BetaAgentNewParams{
		Name:   spec.Name,
		Model:  anthropic.BetaManagedAgentsModelConfigParams{ID: spec.Model},
		System: anthropic.String(spec.System),
		Tools:  []anthropic.BetaAgentNewParamsToolUnion{{OfAgentToolset20260401: toolset()}},
	})
	if err != nil {
		return "", 0, fmt.Errorf("creating agent: %w", err)
	}
	return created.ID, created.Version, nil
}

// toolset enables the built-in tools (bash, files, ...) without approval
// prompts, and keeps web access off.
func toolset() *anthropic.BetaManagedAgentsAgentToolset20260401Params {
	return &anthropic.BetaManagedAgentsAgentToolset20260401Params{
		DefaultConfig: anthropic.BetaManagedAgentsAgentToolsetDefaultConfigParams{
			Enabled: anthropic.Bool(true),
			PermissionPolicy: anthropic.BetaManagedAgentsAgentToolsetDefaultConfigParamsPermissionPolicyUnion{
				OfAlwaysAllow: &anthropic.BetaManagedAgentsAlwaysAllowPolicyParam{},
			},
		},
		Configs: []anthropic.BetaManagedAgentsAgentToolConfigParamsUnion{
			{OfWebFetch: &anthropic.BetaManagedAgentsWebFetchToolConfigParams{Enabled: anthropic.Bool(false)}},
			{OfWebSearch: &anthropic.BetaManagedAgentsWebSearchToolConfigParams{Enabled: anthropic.Bool(false)}},
		},
	}
}

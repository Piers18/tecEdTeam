package chat

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"movie-tracker/internal/tmdb"
	"movie-tracker/internal/watchlist"
)

const systemPrompt = `You are a personal movie and TV series assistant.
You have access to the user's watchlist and viewing history through tools.
Use them to answer questions, provide recommendations, and help the user
decide what to watch next. When making recommendations, consider ratings
the user has given to previously watched content.`

// Agent manages chat completions with function-calling and session history.
type Agent struct {
	client   *openai.Client
	executor *toolExecutor
	sessions *SessionStore
}

func NewAgent(apiKey string, repo *watchlist.Repository, tmdbClient *tmdb.Client) *Agent {
	return &Agent{
		client:   openai.NewClient(option.WithAPIKey(apiKey)),
		executor: newToolExecutor(repo, tmdbClient),
		sessions: NewSessionStore(),
	}
}

// Chat sends a user message, runs tool calls as needed, and returns the final reply.
func (a *Agent) Chat(ctx context.Context, sessionID, userMessage string) (string, error) {
	a.sessions.Append(sessionID, openai.UserMessage(userMessage))

	for {
		history := a.sessions.Get(sessionID)
		messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
		messages = append(messages, openai.SystemMessage(systemPrompt))
		messages = append(messages, history...)

		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.F(openai.ChatModelGPT4oMini),
			Messages: openai.F(messages),
			Tools:    openai.F(toolDefinitions),
		})
		if err != nil {
			return "", fmt.Errorf("openai: %w", err)
		}

		choice := resp.Choices[0]

		if choice.FinishReason == openai.ChatCompletionChoicesFinishReasonToolCalls {
			// Manually build the assistant param since ToParam() is not in this SDK version.
			tcParams := make([]openai.ChatCompletionMessageToolCallParam, len(choice.Message.ToolCalls))
			for i, tc := range choice.Message.ToolCalls {
				tcParams[i] = openai.ChatCompletionMessageToolCallParam{
					ID:   openai.F(tc.ID),
					Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
					Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      openai.F(tc.Function.Name),
						Arguments: openai.F(tc.Function.Arguments),
					}),
				}
			}
			a.sessions.Append(sessionID, openai.ChatCompletionAssistantMessageParam{
				Role:      openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
				ToolCalls: openai.F(tcParams),
			})

			// Execute every requested tool and append results
			for _, tc := range choice.Message.ToolCalls {
				result := a.executor.execute(ctx, tc.Function.Name, tc.Function.Arguments)
				a.sessions.Append(sessionID, openai.ToolMessage(tc.ID, result))
			}
			continue // loop: send results back to model
		}

		// Natural language response
		content := choice.Message.Content
		a.sessions.Append(sessionID, openai.AssistantMessage(content))
		return content, nil
	}
}

// ClearSession deletes a session's history.
func (a *Agent) ClearSession(sessionID string) {
	a.sessions.Delete(sessionID)
}

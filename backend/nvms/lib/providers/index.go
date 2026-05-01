// Package providers provides LLM provider implementations for NVMS.
//
// # Subpackages
//
//	openai - OpenAI API integration
//	anthropic - Anthropic Claude API integration
//	gemini - Google Gemini API integration
//	local - Local LLM integration
//	deepseek - DeepSeek API integration
package providers

// Message represents a chat message in provider requests.
type Message = message

// ChatRequest represents a chat completion request.
type ChatRequest = chatRequest

// Choice represents a chat completion choice.
type Choice = choice

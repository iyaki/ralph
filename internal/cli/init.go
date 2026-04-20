package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iyaki/ralphex/internal/config"
	"github.com/spf13/cobra"
)

var isInteractiveTerminal = func() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

var getWorkingDir = os.Getwd

const defaultInitMaxIterations = 25

const (
	defaultInitAgentName              = "opencode"
	defaultInitSpecsDir               = "specs"
	defaultInitSpecsIndexFile         = "README.md"
	defaultInitImplementationPlanName = "IMPLEMENTATION_PLAN.md"
	defaultInitPromptsDir             = ".ralph/prompts"
	defaultInitLogFile                = "./ralph.log"
	confirmYes                        = "yes"
	confirmNo                         = "no"
)

const (
	questionTypeSelect  = "select"
	questionTypeInput   = "input"
	questionTypeConfirm = "confirm"
)

const (
	questionKeyAgentName              = "agent"
	questionKeyModel                  = "model"
	questionKeyAgentMode              = "agent-mode"
	questionKeyMaxIterations          = "max-iterations"
	questionKeySpecsDir               = "specs-dir"
	questionKeySpecsIndexFile         = "specs-index-file"
	questionKeyImplementationPlanName = "implementation-plan-name"
	questionKeyPromptsDir             = "prompts-dir"
	questionKeyOverwriteExisting      = "overwrite-existing"
	questionKeyEnableLogging          = "enable-logging"
	questionKeyLogFile                = "log-file"
	questionKeyLogTruncate            = "log-truncate"
)

var supportedInitAgents = []string{"opencode", "claude", "cursor"}

var errInvalidConfirmAnswer = errors.New("please answer yes or no")

var errInitValueRequired = errors.New("value cannot be empty")

// InitSession represents one interactive run of ralph init.
type InitSession struct {
	OutputPath          string
	IsTTY               bool
	ExistingConfigFound bool
	Questions           []InitQuestion
	Answers             *InitAnswers
	Confirmed           bool
	Reader              *bufio.Reader
	Writer              io.Writer
}

// InitQuestion represents a single question in the interactive flow.
type InitQuestion struct {
	Key          string
	Prompt       string
	Type         string // "select", "input", "confirm"
	DefaultValue string
	Options      []string
	Required     bool
	Validator    func(string) error
}

// InitAnswers mirrors configuration fields for collection.
type InitAnswers struct {
	AgentName              string
	Model                  string
	AgentMode              string
	MaxIterations          int
	SpecsDir               string
	SpecsIndexFile         string
	ImplementationPlanName string
	PromptsDir             string
	NoLog                  bool
	LogFile                string
	LogTruncate            bool
}

type initAnswerApplier func(*InitAnswers, string) error

var initAnswerAppliers = map[string]initAnswerApplier{
	questionKeyAgentName:              setInitAnswerAgentName,
	questionKeyModel:                  setInitAnswerModel,
	questionKeyAgentMode:              setInitAnswerAgentMode,
	questionKeyMaxIterations:          setInitAnswerMaxIterations,
	questionKeySpecsDir:               setInitAnswerSpecsDir,
	questionKeySpecsIndexFile:         setInitAnswerSpecsIndexFile,
	questionKeyImplementationPlanName: setInitAnswerImplementationPlanName,
	questionKeyPromptsDir:             setInitAnswerPromptsDir,
	questionKeyEnableLogging:          setInitAnswerEnableLogging,
	questionKeyLogFile:                setInitAnswerLogFile,
	questionKeyLogTruncate:            setInitAnswerLogTruncate,
}

// NewInitCommand creates the init command.
func NewInitCommand() *cobra.Command {
	var force bool
	var output string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Ralphex configuration",
		Long:  `Interactive command to generate a ralph.toml configuration file.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeInitCommand(cmd, output, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config without prompt")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Target file path (default: ./ralph.toml)")

	return cmd
}

func executeInitCommand(cmd *cobra.Command, outputPath string, force bool) error {
	session, err := newInitSession(cmd, outputPath)
	if err != nil {
		return err
	}

	shouldContinue, err := prepareInitSession(session, force)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	if err := runInitQuestionnaire(session); err != nil {
		return err
	}

	if err := writeInitConfig(session); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(session.Writer, "Initialized Ralphex configuration at %s\n", session.OutputPath)

	return nil
}

func newInitSession(cmd *cobra.Command, outputPath string) (*InitSession, error) {
	session := &InitSession{
		OutputPath: outputPath,
		Answers:    defaultInitAnswers(),
		Reader:     bufio.NewReader(cmd.InOrStdin()),
		Writer:     cmd.OutOrStdout(),
	}

	session.IsTTY = isInteractiveTerminal()
	if !session.IsTTY {
		return nil, fmt.Errorf("ralph init requires an interactive terminal")
	}

	if session.OutputPath == "" {
		cwd, err := getWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve current working directory: %w", err)
		}

		session.OutputPath = filepath.Join(cwd, "ralph.toml")
	}

	return session, nil
}

func prepareInitSession(session *InitSession, force bool) (bool, error) {
	existingConfig, err := initConfigExists(session.OutputPath)
	if err != nil {
		return false, err
	}
	if !existingConfig {
		return true, nil
	}

	session.ExistingConfigFound = true
	if force {
		return true, nil
	}

	overwriteConfirmed, err := confirmExistingConfigOverwrite(session)
	if err != nil {
		return false, err
	}
	if !overwriteConfirmed {
		_, _ = fmt.Fprintln(session.Writer, "Initialization cancelled; existing configuration was not changed.")

		return false, nil
	}

	return true, nil
}

func initConfigExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("failed to inspect existing configuration at %s: %w", path, err)
	}

	return false, nil
}

func writeInitConfig(session *InitSession) error {
	cfg := buildConfigFromAnswers(session.Answers)
	if err := config.WriteConfig(session.OutputPath, cfg); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

func confirmExistingConfigOverwrite(session *InitSession) (bool, error) {
	answer, err := promptForAnswer(session, newConfirmQuestion(
		questionKeyOverwriteExisting,
		"Overwrite existing configuration?",
		confirmNo,
	))
	if err != nil {
		return false, err
	}

	confirmed, _ := parseConfirmAnswer(answer)

	return confirmed, nil
}

func defaultInitAnswers() *InitAnswers {
	return &InitAnswers{
		AgentName:              defaultInitAgentName,
		MaxIterations:          defaultInitMaxIterations,
		SpecsDir:               defaultInitSpecsDir,
		SpecsIndexFile:         defaultInitSpecsIndexFile,
		ImplementationPlanName: defaultInitImplementationPlanName,
		PromptsDir:             defaultInitPromptsDir,
		NoLog:                  true,
		LogFile:                defaultInitLogFile,
		LogTruncate:            false,
	}
}

func runInitQuestionnaire(session *InitSession) error {
	session.Questions = baseInitQuestions(session.Answers)
	if err := askQuestions(session, session.Questions); err != nil {
		return err
	}

	if session.Answers.NoLog {
		return nil
	}

	loggingQuestions := loggingInitQuestions(session.Answers)
	session.Questions = append(session.Questions, loggingQuestions...)

	if err := askQuestions(session, loggingQuestions); err != nil {
		return err
	}

	return nil
}

func baseInitQuestions(defaults *InitAnswers) []InitQuestion {
	return []InitQuestion{
		newSelectQuestion(
			questionKeyAgentName,
			"AI agent (opencode/claude/cursor)",
			defaults.AgentName,
			supportedInitAgents,
			validateInitAgent,
		),
		newInputQuestion(questionKeyModel, "Model (optional)", defaults.Model, false, nil),
		newInputQuestion(questionKeyAgentMode, "Agent mode/sub-agent (optional)", defaults.AgentMode, false, nil),
		newInputQuestion(
			questionKeyMaxIterations,
			"Maximum iterations",
			strconv.Itoa(defaults.MaxIterations),
			true,
			validatePositiveInitInteger,
		),
		newInputQuestion(questionKeySpecsDir, "Specs directory", defaults.SpecsDir, true, nil),
		newInputQuestion(questionKeySpecsIndexFile, "Specs index file", defaults.SpecsIndexFile, true, nil),
		newInputQuestion(
			questionKeyImplementationPlanName,
			"Implementation plan file",
			defaults.ImplementationPlanName,
			true,
			nil,
		),
		newInputQuestion(questionKeyPromptsDir, "Prompts directory", defaults.PromptsDir, true, nil),
		newConfirmQuestion(questionKeyEnableLogging, "Enable logging?", boolToConfirmValue(!defaults.NoLog)),
	}
}

func loggingInitQuestions(defaults *InitAnswers) []InitQuestion {
	return []InitQuestion{
		newInputQuestion(questionKeyLogFile, "Log file path", defaults.LogFile, true, nil),
		newConfirmQuestion(
			questionKeyLogTruncate,
			"Truncate log file on each run?",
			boolToConfirmValue(defaults.LogTruncate),
		),
	}
}

func newSelectQuestion(
	key, prompt, defaultValue string,
	options []string,
	validator func(string) error,
) InitQuestion {
	return InitQuestion{
		Key:          key,
		Prompt:       prompt,
		Type:         questionTypeSelect,
		DefaultValue: defaultValue,
		Options:      options,
		Required:     true,
		Validator:    validator,
	}
}

func newInputQuestion(
	key, prompt, defaultValue string,
	required bool,
	validator func(string) error,
) InitQuestion {
	return InitQuestion{
		Key:          key,
		Prompt:       prompt,
		Type:         questionTypeInput,
		DefaultValue: defaultValue,
		Required:     required,
		Validator:    validator,
	}
}

func newConfirmQuestion(key, prompt, defaultValue string) InitQuestion {
	return InitQuestion{
		Key:          key,
		Prompt:       prompt,
		Type:         questionTypeConfirm,
		DefaultValue: defaultValue,
	}
}

func askQuestions(session *InitSession, questions []InitQuestion) error {
	for _, question := range questions {
		answer, err := promptForAnswer(session, question)
		if err != nil {
			return err
		}

		if err := applyInitAnswer(session.Answers, question.Key, answer); err != nil {
			return err
		}
	}

	return nil
}

func promptForAnswer(session *InitSession, question InitQuestion) (string, error) {
	for {
		answer, err := askSingleQuestion(session, question)
		if err != nil {
			return "", err
		}

		if validationErr := validateQuestionAnswer(question, answer); validationErr != nil {
			if _, writeErr := fmt.Fprintln(session.Writer, validationErr.Error()); writeErr != nil {
				return "", writeErr
			}

			continue
		}

		return answer, nil
	}
}

func askSingleQuestion(session *InitSession, question InitQuestion) (string, error) {
	if err := printQuestion(session.Writer, question); err != nil {
		return "", err
	}

	answer, err := readAnswer(session.Reader)
	if err != nil {
		return "", normalizeAnswerReadError(err)
	}

	if answer == "" {
		return question.DefaultValue, nil
	}

	return answer, nil
}

func normalizeAnswerReadError(err error) error {
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("unexpected end of input during init questionnaire")
	}

	return err
}

func validateQuestionAnswer(question InitQuestion, answer string) error {
	if question.Required && strings.TrimSpace(answer) == "" {
		return errInitValueRequired
	}

	if question.Type == questionTypeConfirm {
		if _, ok := parseConfirmAnswer(answer); !ok {
			return errInvalidConfirmAnswer
		}

		return nil
	}

	if question.Validator != nil {
		return question.Validator(answer)
	}

	return nil
}

func printQuestion(writer io.Writer, question InitQuestion) error {
	if question.DefaultValue == "" {
		_, err := fmt.Fprintf(writer, "%s: ", question.Prompt)

		return err
	}

	_, err := fmt.Fprintf(writer, "%s [%s]: ", question.Prompt, question.DefaultValue)

	return err
}

func readAnswer(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) {
		if line == "" {
			return "", io.EOF
		}

		return strings.TrimSpace(line), nil
	}
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func applyInitAnswer(answers *InitAnswers, key, value string) error {
	applier, ok := initAnswerAppliers[key]
	if !ok {
		return fmt.Errorf("unknown init question key: %s", key)
	}

	return applier(answers, value)
}

func setInitAnswerAgentName(answers *InitAnswers, value string) error {
	answers.AgentName = value

	return nil
}

func setInitAnswerModel(answers *InitAnswers, value string) error {
	answers.Model = value

	return nil
}

func setInitAnswerAgentMode(answers *InitAnswers, value string) error {
	answers.AgentMode = value

	return nil
}

func setInitAnswerMaxIterations(answers *InitAnswers, value string) error {
	maxIterations, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid max-iterations answer: %w", err)
	}

	answers.MaxIterations = maxIterations

	return nil
}

func setInitAnswerSpecsDir(answers *InitAnswers, value string) error {
	answers.SpecsDir = value

	return nil
}

func setInitAnswerSpecsIndexFile(answers *InitAnswers, value string) error {
	answers.SpecsIndexFile = value

	return nil
}

func setInitAnswerImplementationPlanName(answers *InitAnswers, value string) error {
	answers.ImplementationPlanName = value

	return nil
}

func setInitAnswerPromptsDir(answers *InitAnswers, value string) error {
	answers.PromptsDir = value

	return nil
}

func setInitAnswerEnableLogging(answers *InitAnswers, value string) error {
	enableLogging, ok := parseConfirmAnswer(value)
	if !ok {
		return errInvalidConfirmAnswer
	}

	answers.NoLog = !enableLogging

	return nil
}

func setInitAnswerLogFile(answers *InitAnswers, value string) error {
	answers.LogFile = value

	return nil
}

func setInitAnswerLogTruncate(answers *InitAnswers, value string) error {
	logTruncate, ok := parseConfirmAnswer(value)
	if !ok {
		return errInvalidConfirmAnswer
	}

	answers.LogTruncate = logTruncate

	return nil
}

func parseConfirmAnswer(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes", "true", "1":
		return true, true
	case "n", "no", "false", "0":
		return false, true
	default:
		return false, false
	}
}

func boolToConfirmValue(value bool) string {
	if value {
		return confirmYes
	}

	return confirmNo
}

func validateInitAgent(value string) error {
	for _, agentName := range supportedInitAgents {
		if value == agentName {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", strings.Join(supportedInitAgents, ", "))
}

func validatePositiveInitInteger(value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fmt.Errorf("must be an integer greater than 0")
	}

	return nil
}

func buildConfigFromAnswers(answers *InitAnswers) *config.Config {
	return &config.Config{
		AgentName:              answers.AgentName,
		Model:                  answers.Model,
		AgentMode:              answers.AgentMode,
		MaxIterations:          answers.MaxIterations,
		SpecsDir:               answers.SpecsDir,
		SpecsIndexFile:         answers.SpecsIndexFile,
		ImplementationPlanName: answers.ImplementationPlanName,
		PromptsDir:             answers.PromptsDir,
		NoLog:                  answers.NoLog,
		LogFile:                answers.LogFile,
		LogTruncate:            answers.LogTruncate,
	}
}

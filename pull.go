package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"strings"

	"github.com/phrase/phraseapp-go/phraseapp"
)

type PullCommand struct {
	phraseapp.AuthCredentials
	DebugPull bool `cli:"opt --debug desc='Debug output (only push+pull)'"`
}

func (cmd *PullCommand) Run() error {
	Authenticate(&cmd.AuthCredentials)

	if cmd.DebugPull {
		Debug = true
	}
	targets, err := TargetsFromConfig(cmd)
	if err != nil {
		return err
	}

	for _, target := range targets {
		err := target.Pull()
		if err != nil {
			return err
		}
	}
	return nil
}

type Targets []*Target

type Target struct {
	File          string      `yaml:"file,omitempty"`
	ProjectId     string      `yaml:"project_id,omitempty"`
	AccessToken   string      `yaml:"access_token,omitempty"`
	FileFormat    string      `yaml:"file_format,omitempty"`
	Params        *PullParams `yaml:"params,omitempty"`
	RemoteLocales []*phraseapp.Locale
}

type PullParams struct {
	FileFormat               string                  `yaml:"file_format,omitempty"`
	LocaleId                 string                  `yaml:"locale_id,omitempty"`
	ConvertEmoji             *bool                   `yaml:"convert_emoji,omitempty"`
	FormatOptions            *map[string]interface{} `yaml:"format_options,omitempty"`
	IncludeEmptyTranslations *bool                   `yaml:"include_empty_translations,omitempty"`
	KeepNotranslateTags      *bool                   `yaml:"keep_notranslate_tags,omitempty"`
	Tag                      string                  `yaml:"tag,omitempty"`
}

func (t *Target) GetFormat() string {
	if t.Params != nil && t.Params.FileFormat != "" {
		return t.Params.FileFormat
	}
	if t.FileFormat != "" {
		return t.FileFormat
	}
	return ""
}

func (t *Target) GetLocaleId() string {
	if t.Params != nil {
		return t.Params.LocaleId
	}
	return ""
}

func (t *Target) GetTag() string {
	if t.Params != nil {
		return t.Params.Tag
	}
	return ""
}

func (target *Target) Pull() error {
	if strings.TrimSpace(target.File) == "" {
		return fmt.Errorf("file pattern for target may not be empty")
	}

	pathComponents, err := ExtractPathComponents(target.File)
	if err != nil {
		return err
	}

	target.RemoteLocales, err = RemoteLocales(target.ProjectId)
	if err != nil {
		return err
	}

	localeFile := &LocaleFile{Id: target.GetLocaleId(), Tag: target.GetTag()}
	localeToPathMapping, err := pathComponents.ExpandPathsWithLocale(target.RemoteLocales, localeFile)
	if err != nil {
		return err
	}

	for _, localeToPath := range localeToPathMapping {
		err := createFile(localeToPath.Path)
		if err != nil {
			return err
		}

		err = target.DownloadAndWriteToFile(localeToPath)
		if err != nil {
			return fmt.Errorf("%s for %s", err, localeToPath.Path)
		} else {
			sharedMessage("pull", localeToPath)
		}
		if Debug {
			fmt.Println(strings.Repeat("-", 10))
		}
	}

	return nil
}

func TargetsFromConfig(cmd *PullCommand) (Targets, error) {
	content, err := ConfigContent()
	if err != nil {
		return nil, err
	}

	var config *PullConfig

	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, err
	}

	token := config.Phraseapp.AccessToken
	if cmd.Token != "" {
		token = cmd.Token
	}
	projectId := config.Phraseapp.ProjectId
	fileFormat := config.Phraseapp.FileFormat

	if &config.Phraseapp.Pull == nil || config.Phraseapp.Pull.Targets == nil {
		return nil, fmt.Errorf("no targets for download specified")
	}

	targets := config.Phraseapp.Pull.Targets

	validTargets := []*Target{}
	for _, target := range targets {
		if target == nil {
			continue
		}
		if target.ProjectId == "" {
			target.ProjectId = projectId
		}
		if target.AccessToken == "" {
			target.AccessToken = token
		}
		if target.FileFormat == "" {
			target.FileFormat = fileFormat
		}
		validTargets = append(validTargets, target)
	}

	if len(validTargets) <= 0 {
		return nil, fmt.Errorf("no targets could be identified! Refine the targets list in your config")
	}

	return validTargets, nil
}

func (target *Target) DownloadAndWriteToFile(localeFile *LocaleFile) error {
	downloadParams := target.setDownloadParams()

	params := target.Params
	localeId := ""
	if params != nil && params.LocaleId != "" {
		localeId = params.LocaleId
	} else {
		localeId = localeFile.Id
	}

	if Debug {
		fmt.Println("Target file pattern:", target.File)
		fmt.Println("Actual file path", localeFile.Path)
		fmt.Println("LocaleId", localeId)
		fmt.Println("ProjectId", target.ProjectId)
		fmt.Println("FileFormat", downloadParams.FileFormat)
		fmt.Println("ConvertEmoji", downloadParams.ConvertEmoji)
		fmt.Println("IncludeEmptyTranslations", downloadParams.IncludeEmptyTranslations)
		fmt.Println("KeepNotranslateTags", downloadParams.KeepNotranslateTags)
		fmt.Println("Tag", downloadParams.Tag)
		fmt.Println("FormatOptions", downloadParams.FormatOptions)
	}

	res, err := phraseapp.LocaleDownload(target.ProjectId, localeId, downloadParams)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(localeFile.Path, res, 0700)
	if err != nil {
		return err
	}
	return nil
}

func (target *Target) setDownloadParams() *phraseapp.LocaleDownloadParams {
	downloadParams := new(phraseapp.LocaleDownloadParams)
	downloadParams.FileFormat = target.FileFormat

	params := target.Params

	if target.Params == nil {
		return downloadParams
	}

	format := params.FileFormat
	if format != "" {
		downloadParams.FileFormat = format
	}

	convertEmoji := params.ConvertEmoji
	if convertEmoji != nil {
		downloadParams.ConvertEmoji = convertEmoji
	}

	formatOptions := params.FormatOptions
	if formatOptions != nil {
		downloadParams.FormatOptions = formatOptions
	}

	includeEmptyTranslations := params.IncludeEmptyTranslations
	if includeEmptyTranslations != nil {
		downloadParams.IncludeEmptyTranslations = includeEmptyTranslations
	}

	keepNotranslateTags := params.KeepNotranslateTags
	if keepNotranslateTags != nil {
		downloadParams.KeepNotranslateTags = keepNotranslateTags
	}

	tag := params.Tag
	if tag != "" {
		downloadParams.Tag = &tag
	}

	return downloadParams
}

// Parsing
type PullConfig struct {
	Phraseapp struct {
		AccessToken string `yaml:"access_token"`
		ProjectId   string `yaml:"project_id"`
		FileFormat  string `yaml:"file_format,omitempty"`
		Pull        struct {
			Targets Targets
		}
	}
}

func createFile(path string) error {
	err := exists(path)
	if err != nil {
		absDir := filepath.Dir(path)
		err := exists(absDir)
		if err != nil {
			os.MkdirAll(absDir, 0700)
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func exists(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("no such file or directory:", absPath)
	}
	return nil
}

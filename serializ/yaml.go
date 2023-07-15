package serializ

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"fmt"
)

func YamlToJson(input []byte) ([]byte, error) {
	var buffer map[string]any
	err := yaml.Unmarshal(input, &buffer)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal yaml input ! Caused by: %w", err)
	}
	return json.Marshal(buffer)
}

func YamlToJsonString(input string) (string, error) {
	output, err := YamlToJson([]byte(input))
	return string(output), err
}

func JsonToYaml(input []byte) ([]byte, error) {
	var buffer map[string]any
	err := json.Unmarshal(input, &buffer)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal json input ! Caused by: %w", err)
	}
	return yaml.Marshal(buffer)
}

func JsonToYamlString(input string) (string, error) {
	output, err := JsonToYaml([]byte(input))
	return string(output), err
}
package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/diyan/assimilator/lib/weakconv"
	pickle "github.com/hydrogen18/stalecucumber"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Identifier interface {
	KeyAlias() string
	KeyCanonical() string
	//EncodeResponse() ([]byte, error)
	//EncodeRecord() (interface{}, error)
}

type RequestDecoder interface {
	DecodeRequest(map[string]interface{}) error
}

type RecordDecoder interface {
	DecodeRecord(map[string]interface{}) error
}

type EventDetails struct {
	Ref         int           `kv:"_ref"         in:"-"           json:"-"`
	RefVersion  int           `kv:"_ref_version" in:"-"           json:"-"`
	Server      string        `kv:"server_name"  in:"server_name" json:"-"`
	Logger      string        `kv:"logger"       in:"logger"      json:"-"`
	Level       string        `kv:"level"        in:"level"       json:"-"`
	Culprit     string        `kv:"culprit"      in:"culprit"     json:"-"`
	Platform    string        `kv:"platform"     in:"platform"    json:"-"`
	Release     *string       `kv:"release"      in:"release"     json:"release"`
	Tags        []TagKeyValue `kv:"tags"         in:"tags"        json:"tags"`
	Environment string        `kv:"environment"  in:"environment" json:"-"`
	Fingerprint []string      `kv:"fingerprint"  in:"fingerprint" json:"-"`

	Modules map[string]string      `kv:"modules" in:"modules" json:"packages"`
	Extra   map[string]interface{} `kv:"extra"   in:"extra"   json:"context"`

	// TODO those fields was not mentioned at https://docs.sentry.io/clientdev/attributes/
	Version      string            `kv:"version"  in:"-" json:"-"`
	Type         string            `kv:"type"            json:"type"`
	Size         int               `kv:"-"               json:"size"`
	Errors       []EventError      `kv:"errors"   in:"-" json:"errors"`
	ReceivedTime time.Time         `kv:"received" in:"-" json:"dateReceived"`
	Metadata     map[string]string `kv:"metadata" in:"-" json:"metadata"`
	UserReport   *string           `kv:"-"               json:"userReport"`
}

func TimeDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(time.Time{}) {
		return data, nil
	}
	if timeFloat, ok := data.(float64); ok {
		return time.Unix(int64(timeFloat), 0).UTC(), nil
	} else if timeString, ok := data.(string); ok {
		time, err := time.Parse(time.RFC3339, timeString)
		if err != nil {
			return nil, err
		}
		return time, nil
	}
	return nil, fmt.Errorf("type is neither float64 nor string")
}

func TagsDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf([]TagKeyValue{}) {
		return data, nil
	}
	tags := []TagKeyValue{}
	// Valid tags are both {"tagKey": "tagValue"} and [["tagKey", "tagValue"]]
	if tagsMap, ok := data.(map[string]interface{}); ok {
		for k, v := range tagsMap {
			// TODO check length of tag key and tag value
			tags = append(tags, TagKeyValue{
				Key: weakconv.String(k), Value: weakconv.String(v),
			})
		}
	} else if tagsSlice, ok := data.([]interface{}); ok {
		for _, tagBlob := range tagsSlice {
			// TODO safe type assertion
			tag := tagBlob.([]interface{})
			// TODO check length of tag key and tag value
			tags = append(tags, TagKeyValue{
				Key: weakconv.String(tag[0]), Value: weakconv.String(tag[1]),
			})
		}
	} else {
		return nil, fmt.Errorf("type is neither map[string]interface{} nor []interface{}")
	}
	return tags, nil
}

// TODO Hook works but looks like we have to traverse maps and slices
func PickleNoneDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f != reflect.TypeOf(pickle.PickleNone{}) {
		return data, nil
	}
	return nil, nil
}

func StringMapDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if !(f == reflect.TypeOf(map[interface{}]interface{}{}) &&
		t == reflect.TypeOf(map[string]interface{}{})) {
		return data, nil
	}
	return nil, nil
}

func DecodeRecord(record map[string]interface{}, target interface{}) error {
	metadata := mapstructure.Metadata{}
	decodeHook := mapstructure.ComposeDecodeHookFunc(TimeDecodeHook, TagsDecodeHook, PickleNoneDecodeHook)
	config := mapstructure.DecoderConfig{
		DecodeHook:       decodeHook,
		Metadata:         &metadata,
		WeaklyTypedInput: false,
		TagName:          "kv",
		Result:           target,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return errors.Wrapf(err, "can not decode record from key/value node store")
	}
	err = decoder.Decode(record)
	return errors.Wrapf(err, "can not decode record from key/value node store")
}

func (event *EventDetails) DecodeRecord(record map[string]interface{}) error {
	if err := DecodeRecord(record, event); err != nil {
		return err
	}
	if event.Level == "" {
		event.Level = "error"
	}
	// TODO remove hardcode
	event.Size = 6597
	//event.DateCreated = time.Date(2999, time.January, 1, 0, 0, 0, 0, time.UTC)
	return nil
}

func DecodeRequest(request map[string]interface{}, target interface{}) error {
	metadata := mapstructure.Metadata{}
	decodeHook := mapstructure.ComposeDecodeHookFunc(TimeDecodeHook, TagsDecodeHook)
	config := mapstructure.DecoderConfig{
		DecodeHook:       decodeHook,
		Metadata:         &metadata,
		WeaklyTypedInput: true,
		TagName:          "in",
		Result:           target,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return errors.Wrapf(err, "can not parse request body")
	}
	err = decoder.Decode(request)
	return errors.Wrapf(err, "can not parse request body")
}

type TagKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

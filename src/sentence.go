package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/moraesvic/flashcard-factory/aws"
	"github.com/moraesvic/flashcard-factory/pinyin"
)

type Sentence struct {
	id                     string
	textOriginal           string
	textTransliterated     string
	textTranslated         string
	audioOriginalBytes     []byte
	audioOriginalFile      string
	audioReducedSpeedBytes []byte
	audioReducedSpeedFile  string
	ankiFlashcard          string
}

func createSentenceId(timestamp string, index int) string {
	return fmt.Sprintf("%s-%04d", timestamp, index)
}

func CreateSentence(timestamp string, index int, textOriginal string) Sentence {
	sentence := Sentence{
		id:           createSentenceId(timestamp, index),
		textOriginal: textOriginal,
	}

	return sentence
}

func (s Sentence) ToString() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "ID", s.id))
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "Original text", s.textOriginal))
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "Transliterated text", s.textTransliterated))
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "Translated text", s.textTranslated))
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "Original audio", s.audioOriginalFile))
	sb.WriteString(fmt.Sprintf("%-20s : %-20s\n", "Reduced speed audio", s.audioReducedSpeedFile))

	return sb.String()
}

func (s *Sentence) SynthesizeSpeech(client *polly.Client) {
	s.audioOriginalBytes = aws.SynthesizeSpeech(client, s.textOriginal)
	s.audioOriginalFile = fmt.Sprintf("%s.mp3", s.id)

	err := os.WriteFile(s.audioOriginalFile, s.audioOriginalBytes, 0644)
	if err != nil {
		panic(fmt.Sprintf("Error writing audio data to file: %s", err))
	}
}

func (s *Sentence) Translate(client *translate.Client) {
	s.textTranslated = aws.Translate(client, s.textOriginal)
}

func (s Sentence) Log() {
	log.Println(s.ToString())
}

func (s *Sentence) ToAnkiFlashcard() {
	var translation string

	if len(s.textTranslated) == 0 {
		translation = "(add translation here)"
	} else {
		translation = s.textTranslated
	}

	s.ankiFlashcard = fmt.Sprintf(
		"%s;%s;%s;%s;%s",
		s.textOriginal,
		s.textTransliterated,
		fmt.Sprintf("[sound:%s]", s.audioOriginalFile),
		fmt.Sprintf("[sound:%s]", s.audioReducedSpeedFile),
		translation)
}

func (s *Sentence) ToPinyin() {
	s.textTransliterated = pinyin.ToPinyin(s.textOriginal)
}

func (s *Sentence) ChangeAudioTempo() {
	s.audioReducedSpeedFile = aws.ChangeAudioTempo(s.audioOriginalFile)
}

func (s *Sentence) Process(pollyClient *polly.Client, translateClient *translate.Client) {
	s.SynthesizeSpeech(pollyClient)
	s.ChangeAudioTempo()
	s.ToPinyin()
	s.Translate(translateClient)
	s.ToAnkiFlashcard()
}

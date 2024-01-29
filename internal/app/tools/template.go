package tools

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const path = "templates"

func readFile(path string) ([]string, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("file does not exist")
			return nil, err
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		rows = append(rows, sc.Text())
	}
	return rows, nil

}

func GetDonePaymentTemplate() (string, error) {
	data, err := readFile(filepath.Join("internal", "app", path, "payment_done_page.html"))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return strings.Join(data, "\n"), nil
}

func GetStatusPaymentTemplate() (string, error) {
	data, err := readFile(filepath.Join("internal", "app", path, "payment_status_page.html"))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return strings.Join(data, "\n"), nil
}

func GetStatusPaymentPage(status string) []byte {
	tmp, _ := GetStatusPaymentTemplate()
	return []byte(fmt.Sprintf(tmp, status))
}

func GetMakePaymentTemplate() (string, error) {
	data, err := readFile(filepath.Join("internal", "app", path, "payment_make_page.html"))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return strings.Join(data, "\n"), nil
}

func GetMakePaymentPage() []byte {
	tmp, _ := GetMakePaymentTemplate()
	return []byte(tmp)
}

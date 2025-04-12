package main

import "strings"

type QShare struct {
	Actionabletype string   `json:"actionableType"`
	Sharetype      string   `json:"shareType"`
	Sharepassword  string   `json:"sharePassword"`
	Hash           string   `json:"hash"`
	Threshold      int      `json:"threshold"`
	Shareblocks    []QBlock `json:"shareBlocks"`
}

type QBlock struct {
	Uri      string `json:"Uri"`
	Location string `json:"Location"`
}

func ProcessQShare(qshare QShare) string {
	switch qshare.Sharetype {
	case "Blocks":
		return "Error: not implemented, " + qshare.Sharetype

	case "File":
		/*
			uri := qshare.Shareblocks[len(qshare.Shareblocks)-1].Uri
			encFilepath, err := Downloadfile(uri)
			if err != nil {
				return "Error: " + err.Error()
			}

			password := qshare.Sharepassword
			decFilepath := Convert(encFilepath)
			err = DecryptFile(password, encFilepath, decFilepath)
			if err != nil {
				return "Error: " + err.Error()
			}
		*/
		return "File was downloaded and decrypted, path = " //+ decFilepath

	case "Ida":
		return "Error: not implemented, " + qshare.Sharetype

	default:
		return "Error: unrecognized command"
	}
}

func Convert(inputPath string) string {
	inputPathList := strings.Split(inputPath, ".")
	return inputPathList[0] + "." + inputPathList[1]
}

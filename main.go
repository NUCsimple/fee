package main

import (
	"io/ioutil"
	"fmt"
	"strings"
	"encoding/csv"
	"strconv"
	"flag"
)

type instanceFee struct {
	Duration int
	Fee      float64
	Size     int
}

const (
	medium = 2
	large  = 4
	xlarge = 4
)

var (
	codeIndex         = 6
	payTypeIndex      = 10
	durationIndex     = 14
	instanceIdIndex   = 19
	feeIndex          = 32
	instanceTypeIndex = 24
	dateIndex         = 1
	feeTypeIndex      = 26
)

var (
	feePath = flag.String("file", "", "Please input the path of file.")
)

func main() {
	flag.Parse()

	feeBytes, err := ioutil.ReadFile(*feePath)
	if err != nil {
		fmt.Errorf("failed to parse fee,because of %v", err)
		return
	}
	r := csv.NewReader(strings.NewReader(string(feeBytes)))

	records, err := r.ReadAll()
	if err != nil {
		fmt.Errorf("failed to read all fee,because of %v", err)
		return
	}

	codeIndex, payTypeIndex, durationIndex, instanceIdIndex, feeIndex, instanceTypeIndex, dateIndex, feeTypeIndex = FindIndex(records[0])
	fee := make(map[string]*instanceFee)
	instances := make(map[string]int)

	PreDate := ""

	for index, record := range records {
		if index == 0 {
			// skip title
			continue
		}

		if record[codeIndex] == "ecs" && record[payTypeIndex] == "后付费" && (feeTypeIndex==0 || record[feeTypeIndex] == "云服务器配置") {
			if PreDate == "" {
				PreDate = record[dateIndex]
			} else if PreDate != record[dateIndex] {
				// clean and print
				PrintFee(PreDate, fee)
				fee = make(map[string]*instanceFee)
				instances = make(map[string]int)
				PreDate = record[dateIndex]
				fmt.Printf("\n")
			}

			// 实例与运行时间
			duration, _ := strconv.Atoi(record[durationIndex])

			f, _ := strconv.ParseFloat(record[feeIndex], 32)
			instanceType := record[instanceTypeIndex]

			if iFee := fee[instanceType]; iFee == nil {
				fee[instanceType] = &instanceFee{
					Duration: duration,
					Fee:      f,
					Size:     1,
				}
			} else {

				if instances[record[instanceIdIndex]] == 0 {
					iFee.Size = iFee.Size + 1
				}

				iFee.Duration = iFee.Duration + duration
				iFee.Fee = iFee.Fee + f
			}

			instances[record[instanceIdIndex]] = instances[record[instanceIdIndex]] + duration
		}
	}
	PrintFee(PreDate, fee)
}

func PrintFee(date string, fee map[string]*instanceFee) {
	//fmt.Printf("总台数：%d 总核时：%d 总费用：%f",len(instances),)

	fmt.Printf("日期：%s\n", date)

	var sumFee float64 = 0
	for instanceType, instanceFee := range fee {

		cores := 0

		typeArray := strings.Split(instanceType, ".")
		if len(typeArray) != 3 {
			fmt.Println("Skip InstanceType %s", instanceType)
			cores = 16
		} else {
			size := typeArray[2]
			if size == "medium" {
				cores = 2
			} else if size == "small" {
				cores = 1
			} else if size == "large" {
				cores = 2
			} else {
				a := strings.Split(size, "large")
				if len(a) == 2 {
					if len(a[0]) > 1 {
						b := strings.Split(a[0], "x")

						num, _ := strconv.Atoi(b[0])

						cores = num * 4
					} else {
						cores = 4
					}
				} else {
					cores = 4
				}
			}
		}

		perCoreFee := 3600 * instanceFee.Fee / float64(instanceFee.Duration*cores)
		sumFee += instanceFee.Fee
		fmt.Printf("InstanceType: %30s Amount: %10d Cores: %10d Fee/Core: %10.5f Fee：%10.3f\n", instanceType, instanceFee.Size, cores*instanceFee.Size, perCoreFee, instanceFee.Fee)
	}
	fmt.Printf("按量付费ECS总费用：%f\n", sumFee)
}

func FindIndex(title []string) (codeIndex, payTypeIndex, durationIndex, instanceIdIndex, feeIndex, instanceTypeIndex int, dateIndex int, feeTypeIndex int) {
	for index, txt := range title {
		if txt == "产品Code" {
			codeIndex = index
		}

		if txt == "服务时长" {
			durationIndex = index
		}

		if txt == "消费类型" {
			payTypeIndex = index
		}

		if txt == "实例ID" {
			instanceIdIndex = index
		}

		if txt == "实例规格" {
			instanceTypeIndex = index
		}

		if txt == "应付金额" {
			feeIndex = index
		}

		if txt == "日期" || txt == "消费时间" {
			dateIndex = index
		}

		if txt == "计费项" {
			feeTypeIndex = index
		}
	}
	return
}

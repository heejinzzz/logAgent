package logAgent

import (
	"fmt"
	"github.com/Shopify/sarama"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup
var mutex sync.Mutex // 总日志文件的互斥锁
var Maxsize int
var BackupPath string

func init() {
	Maxsize = 50000                // 总日志文件的最大容量
	BackupPath = "./allLog_backup" // 备份的总日志存放的文件夹
}

// 判断日志文件是否已满
func isFull(file *os.File) bool {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if fileInfo.Size() >= int64(Maxsize) {
		return true
	}
	return false
}

// 将当前的总日志文件备份到备份文件夹中
func backupFile(file *os.File) {
	// 先检查备份文件夹是否存在，若不存在则创建备份文件夹
	_, err := os.Stat(BackupPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(BackupPath, 0777)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	// 用备份时的时间作为备份日志的文件名
	now := time.Now().Format("20060102150405.000000")
	backupName := BackupPath + "\\" + now + ".log"
	newFile, err := os.OpenFile(backupName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	io.Copy(newFile, file)
}

func readLogFromPartition(partitionConsumer sarama.PartitionConsumer, collectedLogFilename string) {
	// 阻塞直到有值发送过来，然后再继续等待
	for msg := range partitionConsumer.Messages() {
		// 从kafka里收到消息后，打开总日志文件，准备写入
		mutex.Lock() // 先给总日志文件上锁
		file, err := os.OpenFile(collectedLogFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			panic(err)
		}

		// 检查总日志文件是否已满
		if isFull(file) {
			backupFile(file)                                                          // 备份总日志
			file, err = os.OpenFile(collectedLogFilename, os.O_RDWR|os.O_TRUNC, 0777) //清空总日志文件
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
		}

		logger := log.New(file, "", 0)
		logger.Println(string(msg.Value))
		mutex.Unlock() // 给总日志文件解锁
		fmt.Println("Receive a new log from kafka:")
		fmt.Println("topic:", msg.Topic, "\tpartition:", msg.Partition, "\toffset:", msg.Offset, "\tcontent:", string(msg.Value))
	}
	defer partitionConsumer.AsyncClose()
	wg.Done()
}

// LogFromKafka 从kafka里读取收集来的日志（消费），将这些日志写入总日志文件（可以自动创建），实现日志收集
func LogFromKafka(kafkaAddress []string, kafkaTopic string, collectedLogFilename string) {
	// 作为消费者连接kafka，创建consumer
	consumer, err := sarama.NewConsumer(kafkaAddress, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// 获取指定的topic下的所有partition
	partitionList, err := consumer.Partitions(kafkaTopic)
	if err != nil {
		panic(err)
	}

	// 遍历每个partition
	for partition := range partitionList {
		// 针对每个partition创建一个相应的分区消费者，其中
		//sarama.OffsetOldest表示从头开始读完整的日志，sarama.OffsetNewest表示从新的日志开始读起
		partitionConsumer, err := consumer.ConsumePartition(kafkaTopic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		// 对每个partition，开启一个goroutine消费其中的消息（取日志）
		wg.Add(1)
		go readLogFromPartition(partitionConsumer, collectedLogFilename)
	}
	wg.Wait()
}

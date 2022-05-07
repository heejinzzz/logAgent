package logAgent

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"time"
)

// LogToKafka 将本机的日志写入kafka
func LogToKafka(logFilename string, kafkaAddress []string, kafkaTopic string) {
	// 创建一个kafka producer
	// 先设置config
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follower都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partitioner
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel中返回
	// 连接kafka，创建一个producer
	client, err := sarama.NewSyncProducer(kafkaAddress, config)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 用tail读日志文件
	// 先设置tailConfig
	tailConfig := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个位置开始读
		MustExist: false,                                // 文件不存在时不报错
		Poll:      true,
	}
	tails, err := tail.TailFile(logFilename, tailConfig) // 读取日志文件
	if err != nil {
		panic(err)
	}

	// 一行一行地往kafka里写日志
	for {
		line, ok := <-tails.Lines // 阻塞直到有值发送过来，然后再继续等待
		if !ok {
			fmt.Println("tail file close ReOpen, filename:", logFilename)
			time.Sleep(time.Second)
			continue
		}

		// 构造要送入kafka的消息
		msg := &sarama.ProducerMessage{}
		msg.Topic = kafkaTopic // 设置消息的topic
		msg.Value = sarama.StringEncoder(line.Text)

		// 向kafka发送msg
		pt, ofs, err := client.SendMessage(msg)
		if err != nil {
			panic(err)
		}
		fmt.Println("A new log has been sent to kafka:")
		fmt.Println("topic:", kafkaTopic, "\tpartition:", pt, "\toffset:", ofs, "\tcontent:", line.Text)
	}
}

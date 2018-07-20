package main

import (
	"crypto/x509"
	"io/ioutil"
	"fmt"
	"crypto/tls"
	"net/http"
	"log"
	"bytes"
	"encoding/json"
	"compress/zlib"
	"encoding/hex"
	"math/rand"
	"time"
)
//定义本地统计上传数据格式
type LocalStatisticsReportMSG struct{
	Head string			`json:"Head"`
	Content string		`json:"Content"`
	ContentId string	`json:"Content_id"`
	Rear string			`json:"Rear"`
}

//获取到的1024个字符串
var getOneKString string
var globalRandomString string

//保存 本地统计获取 服务端返回的数值
var localSaveString string


func main() {

	// 测试: 1、统计配置查询
	//urlStatisticalConfigurationQuery()

	// 测试: 2、本地统计上报
	//uriLocalStatisticsReport()

	// 测试: 3、本地统计获取
	urlLocalStatisticalAcquisition()

	// 测试：4、代报结果提交
	urlSubmissionResults()

	// 测试：5、云统计接口
	//urlCloudInterface()
}

func TLSserver() *http.Client{
	pool := x509.NewCertPool()
	caCertPath := "ca.crt"

	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		log.Println(err)
		return nil
	}
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		log.Println(err)
		return nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cliCrt},
		},
	}
	client := &http.Client{Transport: tr}
	return client
}

/*
 * 随机生成1个字符
 */
func  GetOneRandomString() string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 1; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

/*
 * 随机生成1024个字符
 */
func GetOneKRandomString() string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 1024; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	res := hex.EncodeToString([]byte(result))

	return res
}


/*
 * 1、统计部署查询
 */
func urlStatisticalConfigurationQuery()  {

	client := TLSserver()

	req, err := http.NewRequest("GET","https://localhost:7070/v3/api/jgstatisc/collect.cfg?devType=0&net=0&devId=abc",nil)
	if err != nil {
		log.Print("New request error")
		log.Println("err")
	}

	response, err := client.Do(req)
	defer response.Body.Close()

	if err != nil {
		log.Print("do request error")
		log.Println(err)
	}
	fmt.Println("HttpRespStatisConfigQuery succeed")
}

/*
 * 2、本地统计上报
 */
func uriLocalStatisticsReport(){

	client := TLSserver()

	//生成随机字符
	singleString := GetOneRandomString()
	fmt.Println("singleString:"+singleString)

	//构造输出数据项
	var lsrMSG LocalStatisticsReportMSG
	lsrMSG.Head = "{"
	lsrMSG.Content = "测试：用户数据项:"
	lsrMSG.ContentId = singleString
	lsrMSG.Rear = "}"

	//将go类结构体转化为json
	lsrByte , errJson := json.Marshal(lsrMSG)
	if errJson == nil {
		var lsrMSG LocalStatisticsReportMSG
		if err := json.Unmarshal(lsrByte, &lsrMSG); err == nil {
			fmt.Print("lsrMSG:")
			fmt.Println(lsrMSG)
		} else {
			log.Println(err)
		}
	}else{
		fmt.Println(errJson)
	}

	//压缩格式
	var b bytes.Buffer
	writerTemp := zlib.NewWriter(&b)
	writerTemp.Write([]byte(lsrByte))
	fmt.Print("WriterTemp:")
	fmt.Println(writerTemp)
	//log.Println(writerTemp)
	writerTemp.Close()


	body := bytes.NewReader(b.Bytes())

	fmt.Print("Body:")
	fmt.Println(body)
	req, err := http.NewRequest("POST","https://localhost:7070/v3/api/jgstatisc/collect.do?devId=abc", body)
	if err != nil {
		log.Println(err)
		log.Println(req)
	}
	fmt.Println("NewRequest succeed")

	response, err := client.Do(req)
	if response != nil {
		fmt.Println(response)
	}
	defer response.Body.Close()
	fmt.Println("HttpRespStatisConfigQuery succeed")
}

/*
 * 3、本地统计获取
 */
func urlLocalStatisticalAcquisition (){

	client := TLSserver()

	req, err:= http.NewRequest("GET", "https://localhost:7070/v3/api/jgstatisc/collect.get?devType=0&net=0&devId=abc",nil)
	if err != nil {
		log.Println("get requset error:")
		log.Println(err)
	}

	response, err := client.Do(req)

	temp , err := ioutil.ReadAll(response.Body)
	if err!=nil {
		log.Printf("response body read err",temp)
	}
	localSaveString = string(temp)

	defer response.Body.Close()
	if err != nil {
		log.Println("do request error:")
		log.Println(err)
	}
	fmt.Println("LocalStatisticalAcquisition succeed")


}

/*
 * 4、代报结果提交
 */
func urlSubmissionResults(){

	client := TLSserver()

	headDate := localSaveString[:16]
	fmt.Printf("headData:%s \n",headDate)
	hexString := hex.EncodeToString([]byte(headDate))
	fmt.Printf("hexString:%s \n",hexString)
	body := bytes.NewReader([]byte(hexString))
	fmt.Printf("body:%s \n",body)

	req, err := http.NewRequest("GET","https://localhost:7070/v3/api/jgstatisc/collect.del?head=abcd",body)
	if err != nil {
		log.Print("New request error")
		log.Println("err")
	}
	response, err := client.Do(req)
	defer response.Body.Close()
	if err != nil {
		log.Print("do request error")
		log.Println(err)
	}

	log.Printf("log:1024 随机字符前16字节的hex string:%s", hexString)

	fmt.Sprintf("fmt:1024 随机字符前16字节的hex string:%s", hexString)
	fmt.Println("Submission Results succeed")
}

/*
 * 5、云统计接口
 */
func urlCloudInterface(){

	client := TLSserver()

	testString := GetOneKRandomString()
	body := bytes.NewReader([]byte(testString))
	log.Printf("body:%s",body)

	req, err := http.NewRequest("POST","https://localhost:7070/v3/version.new",body)
	if err != nil {
		log.Print("New request err")
		log.Println(err)
	}
	//response, _ :=client.Do(req)
	response, err := client.Do(req)
	if err != nil {
		log.Println("1")
		log.Println(err)
		return
	}
	defer response.Body.Close()
	if err !=nil{
		log.Print("Do request err")
		log.Println(err)
	}
	log.Print("CloudInterface succeed")
}

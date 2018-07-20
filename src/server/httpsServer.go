package main

import (
    "io/ioutil"
    "net/http"
    "fmt"
    "crypto/x509"
    "crypto/tls"
    "log"
    "math/rand"
    "encoding/json"
    "bytes"
    "compress/zlib"
    "io"
    "os"
    "time"
    "encoding/hex"
)

const STATUS_GET_RESPONSE_OK = 200
const STATUS_GET_RESPONSE_ERROR = 201
const STATUS_POST_REQUEST_OK = 202
const STATUS_POST_REQUEST_ERROR = 204
const ERROR_FOR_UNKNOWN = 100

//Json格式设计的内部类
type status struct {
    Code    int    `json:Code`
    Message string `json:Message`
}

//定义Json格式类
type commitMSG struct {
    Status  status `json:"Status"`
    Collect bool   `json:Collect`
    Commit  bool   `json:Commit`
}

//定义本地统计上传数据格式
type LocalStatisticsReportMSG struct {
    Head      string `json:"Head"`
    Content   string `json:"Content"`
    ContentId string `json:"Content_id"`
    Rear      string `json:"Rear"`
}

//定义1024长度的随机字符串
var global_oneKRandomString string


type myhandler struct {
}

func (h *myhandler) ServeHTTP(w http.ResponseWriter,
   r *http.Request) {
        fmt.Fprintf(w,
       "Hi, This is an example of http service in golang!\n")
}

func main() {
    pool := x509.NewCertPool()
    caCertPath := "ca.crt"

    caCrt, err := ioutil.ReadFile(caCertPath)
    if err != nil {
       fmt.Println("ReadFile err:", err)
       return
    }
    pool.AppendCertsFromPEM(caCrt)

    s := &http.Server{
       Addr:    ":7070",
       //Handler: &myhandler{},
       TLSConfig: &tls.Config{
           ClientCAs:  pool,
           ClientAuth: tls.RequireAndVerifyClientCert,
       },
    }

    //url指定绑定：统计配置查询
    http.HandleFunc("/v3/api/jgstatisc/collect.cfg",RespStatisConfigQuery)

    //url指定绑定：本地统计上报
    http.HandleFunc("/v3/api/jgstatisc/collect.do", LocalStatisticsReport)

    //url指定绑定：本地统计获取
    http.HandleFunc("/v3/api/jgstatisc/collect.get",LocalStatisicsacquisition)

    //url指定绑定：代报结果提交
    http.HandleFunc("/v3/api/jgstatisc/collect.del",SubmissionResults)

    //url指定绑定：云服务接口
    http.HandleFunc("/v3/version.new",CloudStatisticsInterface)

    err = s.ListenAndServeTLS("server.crt", "server.key")
    if err != nil {
       fmt.Println("ListenAndServeTLS err:", err)
    }
}


/*
 * 随机数生成commit值
 */
func getCommitResult() bool {
    var res bool
    temp := rand.Intn(100)
    if temp%2 == 1 {
        res = false
    } else {
        res = true
    }
    return res
}

/*
 * 1、统计配置查询
 */
func RespStatisConfigQuery(w http.ResponseWriter, r *http.Request) {
    fmt.Println("test httpsRespStatisConfigQuery begin...")
    //测试路径请求发送路径
    //fmt.Println(r.RequestURI)

    //获取随机commit数值
    commitRst := getCommitResult()

    //对Json数据赋值
    var msg commitMSG
    msg.Status.Code = 0
    msg.Status.Message = "OK"
    msg.Collect = true
    msg.Commit = commitRst

    //将go类结构体转化为json
    bs, err := json.Marshal(msg)
    if err == nil {
        var respMSG commitMSG
        if err := json.Unmarshal(bs, &respMSG); err == nil {
            log.Print("respMSG:")
            log.Println(respMSG)
        } else {
            fmt.Println(err)
        }

        w.WriteHeader(STATUS_GET_RESPONSE_OK)
        w.Write(bs) //没有显示信息原因:没有吧数据缓存写回 w .
    } else {
        fmt.Print("Json Marshal err:")
        fmt.Println(err)
    }

    //输出请求
    log.Println("request:")
    log.Println(r)

    //输出参数
    r.ParseForm()
    devType := r.FormValue("devtype")
    net := r.FormValue("net")
    devId := r.FormValue("devId")
    log.Printf("devType:%s net:%s devId:%s", devType, net, devId)
    fmt.Sprintf("参数：devType:%s net:%s devId:%s", devType, net, devId)
}

/*
 * 2、本地统计上报
 */
func LocalStatisticsReport(w http.ResponseWriter, r *http.Request) {

    r.ParseForm()
    body, _ := ioutil.ReadAll(r.Body)

    b := bytes.NewReader(body)
    rz, err := zlib.NewReader(b)
    if err != nil {
        log.Println(err)
    }
    io.Copy(os.Stdout, rz)

    w.WriteHeader(STATUS_POST_REQUEST_OK) //返回状态码202

    //输出解压后的数据
    fmt.Print("fmt:rz:")
    fmt.Println(rz)
    fmt.Print("log:rz:")
    log.Println(rz)

    //输出请求
    log.Println("request:")
    log.Println(r)

    //输出参数
    r.ParseForm()
    devId := r.FormValue("devId")
    log.Printf("参数：devId:%s",devId)
    fmt.Sprintf("参数：devId:%s",devId)
}

/*
 * 3、本地统计获取
 */
func LocalStatisicsacquisition(w http.ResponseWriter, r *http.Request) {

    //测试
    fmt.Println("start LocalStatisicsacquisition")

    global_oneKRandomString = GetOneKRandomString()

    w.WriteHeader(STATUS_POST_REQUEST_OK)	//返回状态码202
    w.Write([]byte(global_oneKRandomString))

    //输出请求
    log.Println("request:")
    log.Println(r)

    //输出本地生成的数据
    log.Println("Random string :")
    log.Println(global_oneKRandomString)

    r.ParseForm()

    //输出参数
    devType := r.FormValue("devType")
    net := r.FormValue("net")
    devId := r.FormValue("devId")
    log.Printf("log:参数：devType:%s net:%s devId:%s",devType, net, devId)
    fmt.Sprintf("fmt:参数：devType:%s net:%s devId:%s",devType, net, devId)
}

/*
 * 4、代报结果提交
 */
func SubmissionResults(w http.ResponseWriter, r *http.Request){

    body, _ := ioutil.ReadAll(r.Body)
    w.Write(body)
    w.WriteHeader(STATUS_POST_REQUEST_OK)	//返回状态码202

    r.ParseForm()
    head :=r.FormValue("head")

    //输出请求
    log.Println("request:")
    log.Println(r)

    //输出参数
    log.Printf("log:参数：head:%s ",head)
    fmt.Sprintf("fmt:参数：head:%s ",head)
}

/*
 * 5、有度云统计接口
 */
func CloudStatisticsInterface(w http.ResponseWriter, r *http.Request){

    r.ParseForm()
    body, _ := ioutil.ReadAll(r.Body)

    w.Write([]byte(body))					//显示响应的数据
    w.WriteHeader(STATUS_POST_REQUEST_OK)	//返回状态码202

    //输出请求
    log.Println("request:")
    log.Println(r)

    //输出用户发送数据
    log.Println("respone body:")
    log.Println(body)
}

/*
 * 随机生成1个字符
 */
func GetOneRandomString() string {
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





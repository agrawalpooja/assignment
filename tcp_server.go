package main 
import "net"
import "fmt"
import "io/ioutil"
import "bytes"
import "strconv"
import "os"
import "time"
type fileinfo struct{
	version int
	numBytes int
	creationTime int
	expTime int
}
var m map[string]fileinfo
func writeFile(filename string, numBytes int, expTimeStr int, content []byte,conn net.Conn){
	//fmt.Printf("%v : %v : %v : %v ",filename,numBytes,expTimeStr,content)
	elem, ok := m[filename]
	if ok==true{
		ctime:=time.Now().Unix()
		if (elem.creationTime + elem.expTime) < int(ctime){
			ioutil.WriteFile(filename,content,0644)
			m[filename] = fileinfo{1, numBytes, int(ctime), expTimeStr}
		}else{
			ioutil.WriteFile(filename,content,0644)
			m[filename] = fileinfo{elem.version+1, numBytes, int(ctime), expTimeStr }
		}
	}else{
		ioutil.WriteFile(filename,content,0644)
		ctime:=time.Now().Unix()
		m[filename] = fileinfo{1, numBytes, int(ctime), expTimeStr}
	}
	Finfo := m[filename]
	//fmt.Println("Ok "+strconv.Itoa(Finfo.version))
	fmt.Fprintf(conn, "OK %v\r\n",Finfo.version)
	//conn.Write([]byte("Ok "+strconv.Itoa(Finfo.version)+"\r\n"))
}
func casFile(filename string, version int, numBytes int, expTimeStr int, content []byte,conn net.Conn){
	//fmt.Printf("%v : %v : %v : %v ",filename,numBytes,expTimeStr,content)
	elem, ok := m[filename]
	if ok==true{
		ctime:=time.Now().Unix()
		if (elem.creationTime + elem.expTime) > int(ctime){
			if elem.version==version{
				m[filename] = fileinfo{elem.version, numBytes, int(ctime), expTimeStr }
				ioutil.WriteFile(filename,content,0644)
				Finfo := m[filename]
				fmt.Fprintf(conn, "OK %v\r\n",Finfo.version)
			}else{
				fmt.Fprintf(conn, "ERR_VERSION %v\r\n", elem.version)
			}
		}else{
			m[filename] = fileinfo{1, numBytes, int(ctime), expTimeStr}
			ioutil.WriteFile(filename,content,0644)
			Finfo := m[filename]
			fmt.Fprintf(conn, "OK %v\r\n",Finfo.version)
		}	
	}else{
		ctime:=time.Now().Unix()
		m[filename] = fileinfo{1, numBytes, int(ctime), expTimeStr}
		ioutil.WriteFile(filename,content,0644)
		Finfo := m[filename]
		fmt.Fprintf(conn, "OK %v\r\n",Finfo.version)
	}
	//fmt.Println("Ok "+strconv.Itoa(Finfo.version))
	
	//conn.Write([]byte("Ok "+strconv.Itoa(Finfo.version)+"\r\n"))
}
func readFile(filename string, conn net.Conn){
	elem, ok := m[filename]
	if ok==true{
		ctime:=time.Now().Unix()
		if (elem.creationTime + elem.expTime) > int(ctime){
			content,_ := ioutil.ReadFile(filename)
			newexp := (elem.creationTime + elem.expTime) - (int(ctime))
			fmt.Fprintf(conn, "CONTENTS %v %v %v \r\n%v\r\n", elem.version, elem.numBytes, newexp, string(content))
		//conn.Write([]byte("CONTENTS "+strconv.Itoa(elem.version)+" "+strconv.Itoa(elem.numBytes)+" "+strconv.Itoa(elem.expTime)+" \r\n"+string(content)+"\r\n"))
		//fmt.Println(string(content))
		}else{
			fmt.Fprintf(conn, "ERR_FILE_EXPIRED\r\n")
			deleteFile(filename,conn)
		}
	}else{
		fmt.Fprintf(conn, "ERR_FILE_NOT_FOUND\r\n")
	}
}
func deleteFile(filename string, conn net.Conn){
	_, ok := m[filename]
	if ok==true{
		os.Remove(filename)
		delete(m,filename)
		fmt.Fprintf(conn, "OK\r\n")
	}else{
		fmt.Fprintf(conn, "ERR_FILE_NOT_FOUND\r\n")
	}
}
func parseCommand(command []byte,conn net.Conn){
	//var strArr []string
	//fmt.Println("1st split")
	commArr := bytes.SplitN(command,[]byte("\r\n"),2)
	//fmt.Print(string(commArr[0]))
	//fmt.Println("2nd split")
	strArr := bytes.Split(commArr[0],[]byte{' '})
	//fmt.Println("Done")
	if string(strArr[0])=="write" {
		expT := 9999999999
		if len(strArr)>3{
			expT,_=strconv.Atoi(string(strArr[3]))
		}
		numBytes,_:=strconv.Atoi(string(strArr[2]))
		content := commArr[1][:numBytes]
		//fmt.Println(string(content))
		//fmt.Println(string(commArr[1][numBytes:numBytes+2]))
		if string(commArr[1][numBytes:numBytes+2])=="\r\n"{
			writeFile(string(strArr[1]),numBytes,expT,content,conn)
			remComm := commArr[1][numBytes+2:]
			//fmt.Print(string(remComm))
			//fmt.Printf("%v",len(remComm))
			if remComm[0]!=0{
				parseCommand(remComm,conn)
			}
		}else{
			fmt.Fprintf(conn, "ERR_INTERNAL\r\n")
		}
		//fmt.Println("Done with parsing")	
	}else if string(strArr[0])=="read" {
		if strArr[1][0]==0{
			fmt.Fprintf(conn, "ERR_CMD_ERR\r\n")
		}else{
			//fmt.Println(string(strArr[1]))
			readFile(string(strArr[1]),conn)
			if commArr[1][0]!=0{
				parseCommand(commArr[1],conn)
			}
		}

	}else if string(strArr[0])=="cas" {
		expT := 9999999999
		if len(strArr)>4{
			expT,_=strconv.Atoi(string(strArr[4]))
		}
		version,_:=strconv.Atoi(string(strArr[2]))
		numBytes,_:=strconv.Atoi(string(strArr[3]))
		content := commArr[1][:numBytes]
		//fmt.Println(string(content))
		//fmt.Println(string(commArr[1][numBytes:numBytes+2]))
		if string(commArr[1][numBytes:numBytes+2])=="\r\n"{
			casFile(string(strArr[1]),version,numBytes,expT,content,conn)
			remComm := commArr[1][numBytes+2:]
			//fmt.Print(string(remComm))
			//fmt.Printf("%v",len(remComm))
			if remComm[0]!=0{
				parseCommand(remComm,conn)
			}
		}else{
			fmt.Fprintf(conn, "ERR_INTERNAL\r\n")
		}
		//fmt.Println("Done with parsing")	
	}else if string(strArr[0])=="delete" {
		if strArr[1][0]==0{
			fmt.Fprintf(conn, "ERR_CMD_ERR\r\n")
		}else{
			//fmt.Println(string(strArr[1]))

			deleteFile(string(strArr[1]),conn)
			if commArr[1][0]!=0{
				parseCommand(commArr[1],conn)
			}
		}

	}else{
		fmt.Fprintf(conn, "ERR_CMD_ERR\r\n")
	}
}
func handleClient(conn net.Conn){
	for {      
			message := make([]byte, 1024) 
			conn.Read(message) 
			parseCommand(message,conn)
		} 
}

func serverMain() {
	fmt.Println("Server is ready....")   // listen on all interfaces
	m = make(map[string]fileinfo)
	ln, _ := net.Listen("tcp", ":8080")  
	 // accept connection on port
	for{
		conn, err:= ln.Accept()   
		if err!=nil{
			fmt.Println(err)
		}
		go handleClient(conn)
	}
}
func main() {
	serverMain()
}
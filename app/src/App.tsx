import { useEffect, useState } from 'react'
import './App.css'
import axios from "axios"
import 'semantic-ui-css/semantic.min.css'
import { 
  Button, Container, 
  Card, CardHeader, CardContent, 
  Grid, GridRow, GridColumn, 
  SegmentGroup, Segment, 
  DropdownMenu, DropdownItem, Dropdown,
  Message,
} from 'semantic-ui-react'
import { UnControlled as CodeMirror } from "react-codemirror2";
import "codemirror/lib/codemirror.css";
import "codemirror/lib/codemirror.js";
import 'codemirror/mode/go/go.js';
import 'codemirror/theme/base16-light.css';
import 'codemirror/addon/display/fullscreen.css';
import 'codemirror/addon/edit/matchbrackets.js';
import 'codemirror/addon/selection/active-line.js';
import 'codemirror/addon/fold/foldgutter.css';
import 'codemirror/addon/fold/foldcode.js';
import 'codemirror/addon/fold/foldgutter.js';
import 'codemirror/addon/fold/brace-fold.js';
import 'codemirror/addon/fold/comment-fold.js';

function App() {

  // output
  const [message, setMessage] = useState('')
  const [type, setType] = useState('')
  const renderMessage = () => {
    if(type === 'success') {
      return (
        <Message success>{
          message.split('\n').map((line, i) => {
            return (
              <p key={i}>{line}</p>
            )
          })
        }</Message>
      )
    } else if(type === 'error') {
      return (
        <Message error>{
          message.split('\n').map((line, i) => {
            if(i != 0) return (
              <p key={i}>{line}</p>
            )
          })
        }</Message>
      )
    } else if(type === 'timeout') {
      return (
        <Message warning>{message}</Message>
      )
    }
  }

  // examples
  const example = [
    `package main
import (
    "fmt"
)
func main() {
    fmt.Println("Hello, World!")
    fmt.Println("Hello, World2!")
}`,
`package main
import (
    "github.com/gin-gonic/gin"
)
func main() {
    fmt.Println("Hello, World!")
}`,
`package main
import (
    "fmt"
)
func main() {
	for i := 1; i < 1111111111; i++ {
		fmt.Println(i)
	}
}`
  ]

  // loading control
  const [runLoading, setRunLoading] = useState(false)
  const [saveLoading, setSaveLoading] = useState(false)

  // code
  const [code, setCode] = useState<string | undefined>();
  const runCode = ()=>{
    if(runLoading||saveLoading) return false
    setRunLoading(true)
    axios.post('/api/run',{
      content: code
    }).then(
        response => {
          const {message, type} = response.data
          setMessage(message)
          setType(type)
          searchCode()
        }
    ).finally(()=>{
      setRunLoading(false)
    })
  }
  const saveCode = ()=>{
    if(runLoading||saveLoading) return false
    setSaveLoading(true)
    const postData : {
      content: string | undefined,
      id?: string | undefined
    } = {
      content: code
    }
    let uri = '/api/save'
    if(active) {
      uri = '/api/update'
      postData.id = active
    }
    axios.post(uri,postData).then(
        () => {
          searchCode()
        }
    ).finally(()=>{
      setSaveLoading(false)
    })
  }

  // codeList
  const [codeList, setCodeList] = useState([]);
  const searchCode = ()=>{
    axios.get('/api/search').then(
        response => {
          setCodeList(response.data||[])
        }
    )
  }

  // default active selection
  const [active, setActive] = useState("");
  const clickSegment = (active : string, content : string, type ?: string) => {
    setActive(active)
    setCode(content)
    if(type == 'save') {
      setSaveText('Update')
      setVisibility('visible')
    } else if(type == 'exec') {
      setVisibility('hidden')
    } else {
      setSaveText('Save')
      setVisibility('visible')
    }
    setMessage('')
  }

  // save button control
  const [visibility, setVisibility] = useState("visible")
  const [saveText, setSaveText] = useState("Save")


  useEffect(()=>{
    searchCode()
  }, []);

  return (
    <>
      <Container>
      <Grid>
        <GridRow columns={2}>
          <GridColumn width={4}>
            <Card className='w-full'>
              <CardContent>
                <CardHeader className="left font-one-half">Saved Code List
                
                </CardHeader>
              </CardContent>
              <CardContent className='code-list'>
                <SegmentGroup>
                  <Segment className={'left ' +  (active == "" ? "colde-focus" : "")} onClick={() => {
                    clickSegment("", "")
                  }}>Default</Segment>
                  {
                    codeList.map((item : {type:string, id:string, name:string, content: string, time: string},index)=>{
                      if(item.type == "save")
                        return <Segment className={'left ' +  (active == item.id ? "colde-focus" : "")} onClick={() => {
                          clickSegment(item.id, item.content, item.type)
                        }} key={index}>{item.name.split('.')[0]}</Segment>
                    })
                  }
                </SegmentGroup>
                </CardContent>
            </Card>
            <Card className='w-full'>
              <CardContent>
                <CardHeader className="left font-one-half">Exec Code List
                
                </CardHeader>
              </CardContent>
              <CardContent className='code-list'>
                <SegmentGroup>
                  {
                    codeList.map((item : {type:string, id:string, name:string, content: string, time: string},index)=>{
                      if(item.type == "exec")
                        return <Segment className={'left ' +  (active == item.id ? "colde-focus" : "")} onClick={() => {
                          clickSegment(item.id, item.content, item.type)
                        }} key={index}>{item.name.split('.')[0] + '(' + item.time + ')'}</Segment>
                    })
                  }
                </SegmentGroup>
                </CardContent>
            </Card>
          </GridColumn>
          <GridColumn width={12}>
            <Card className='w-full'>
              <CardContent>
                <CardHeader className="left font-one-half">Code
                <Dropdown text='Examples' className='code-examples'>
                    <DropdownMenu>
                      <DropdownItem text='Example1' onClick={()=>{
                        setCode(example[0])
                      }} description='run successfully' />
                      <DropdownItem text='Example2' onClick={()=>{
                        setCode(example[1])
                      }} description='run failed' />
                      <DropdownItem text='Example3' onClick={()=>{
                        setCode(example[2])
                      }} description='run timeout' />
                    </DropdownMenu>
                  </Dropdown>
                <Button className='pull-right' size='mini' style={{visibility: visibility}} primary loading={saveLoading} onClick={saveCode}>{saveText}</Button>
                <Button className='pull-right' size='mini' positive loading={runLoading} onClick={runCode}>Run</Button>
                </CardHeader>
              </CardContent>
              <CardContent className='p-0'>
                <CodeMirror
                  className='mirror'
                  value={code}
                  options={{
                    lineNumbers: true,
                    mode: "go",
                    theme: "base16-light",
                    extraKeys: { Ctrl: "autocomplete" },
                    autofocus: true,
                    styleActiveLine: true,
                    lineWrapping: true,
                    foldGutter: true,
                    matchBrackets: true,
                    fullscreen: true,
                    lint: true,
                    showCursroWhenSelecting: true,
                    gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"],
                  }}
                  onBlur={(editor) => {
                    setCode(editor.getValue());
                  }}
                />
              </CardContent>
            </Card>
            <Card className='w-full'>
              <CardContent>
                <CardHeader className="left font-one-half">Output
                </CardHeader>
              </CardContent>
              <CardContent className='p-0 code-output left'>
                { renderMessage() }
              </CardContent>
            </Card>
          </GridColumn>
        </GridRow>
      </Grid>
      </Container>
    </>
  )
}

export default App

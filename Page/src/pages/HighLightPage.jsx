import {React,useState,useEffect} from 'react';

import {Select,SelectItem}  from '@heroui/react'

import axios from 'axios'
import ActionTable from "../components/ActionTable";

function HighLightPage(props) {
    const [data,setData] = useState([])

    const [key,setKey] = useState({})
    useEffect(() => {
       axios.get("/api/highlight").then((res) => {
           var tmp  = []
           Object.keys(res.data.data).forEach((key) => {
               tmp.push({
                   'keyword':key,
                   'msg':res.data.data[key]
               })
           })
           setData(tmp)
           if (tmp.length > 0) {
               setKey(tmp[0])
           }
       })
    },[])
    return (
        <div>
            {data.length && <div>
                <Select className={'max-w-xs mb-2'} defaultSelectedKeys={[data[0].keyword]}>
                    {data.map((item) => {
                        return (
                            <SelectItem key={item.keyword} onClick={() => {
                                setKey(item)
                            }}>
                                {item.keyword}
                            </SelectItem>
                        )
                    })}
                </Select>
                <ActionTable dataSource={key.msg} handlePageChange={(page0, pageSize) => {

                }} total={key.length} />
            </div>

            }


        </div>
    );
}

export default HighLightPage;
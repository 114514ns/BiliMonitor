import React, {useState} from 'react';
import {Button, Input, ToastProvider, Image, Pagination, addToast} from "@heroui/react";
import axios from "axios";
import DynamicCard from "../components/DynamicCard";





function ReactionPage(props) {

    const [data,setData] = useState([])
    const [uid,setUid] = useState('')
    const [ups,setUps] = useState([])
    const [select,setSelect] = useState({})
    const [filtered,setFiltered] = useState([])

    const [page,setPage] = useState(1)
    const SIZE = 50

    const listRef = React.createRef();

    return (
        <div>
            <ToastProvider placement={'top-right'}/>
            <div className={'flex flex-row items-center'}>
                <Input label={'UID'} className={'max-w-xs'} value={uid} onValueChange={(e) => {
                    setUid(e)
                }}></Input>
                <Button className={'ml-2'} onClick={() => {
                    axios.get("/api/reaction?mid=" + uid).then(res =>{
                        setData(res.data.data)
                        setFiltered(res.data.data)
                        var map = new Map()
                        if (res.data.data == null) {
                            addToast({
                                title: "Result",
                                description: "No Results",
                                color: 'danger',

                            })
                            return
                        }
                        res.data.data.forEach((item,i)=>{
                            map.set(item.TargetName,item.TargetID)
                        })
                        var dst = []
                        dst.push({
                            UName:'Any',
                            UID :-1
                        })
                        map.forEach((item,i)=>{
                            dst.push({
                                UName:i,
                                UID :item
                            })
                        })
                        setUps(dst)
                        console.log(res.data.data)
                    })
                }}>Search</Button>
            </div>
            <div className={'flex flex-col mt-4'}>
                {              <div className="flex flex-row mr-4 max-h-[85vh] overflow-x-auto overflow-y-hidden w-full">
                    {(ups).map((key)=>{
                        return (
                            <div className={'flex flex-col items-center  hover:bg-[#F3F4F5] min-w-[64px]'} onClick={(e) => {
                                if (key.UID === -1) {
                                    setFiltered(data)
                                }
                                var t = data.filter((item) => item.TargetID === key.UID)
                                setFiltered(t)
                            }}>
                                <img
                                    src={`${AVATAR_API}${key.UID}`}
                                    className="w-[48px] h-[48px]"
                                    style={{ borderRadius: '50%' }}
                                />
                                {!isMobile()?                                <p
                                    className="text-center text-sm truncate w-full hover:text-[#00a1d6]"
                                    title={key.UName}
                                >{key.UName}</p>:<></>}
                            </div>
                        )
                    })}

                </div>}
                <div className={'overflow-y-auto max-h-[85vh] overflow-x-hidden'} ref={listRef}>
                    {(filtered??[]).slice((page-1)*SIZE,page*SIZE).map((item,i)=>{
                        return (
                            <DynamicCard item={item} onClick={() => {
                                window.open(`https://t.bilibili.com/${item.OID}`)
                            }}/>
                        )
                    })}
                    <Pagination initialPage={1} total={Math.ceil((filtered??[]).length/SIZE)} onChange={(e) => {
                        setPage(e)
                        listRef.current.scrollTo({
                            top: 0,
                            behavior: "smooth"
                        })}}></Pagination>
                </div>
                <div>

                    {!isMobile() && select.OID &&  <div  className={'w-[70vw] h-full'}>
                        {select.Type === "DYNAMIC_TYPE_AV" && <iframe src={'https://www.bilibili.com/video/' + select.BV} className={'w-full h-full'}/>}
                        {select.Type !== "DYNAMIC_TYPE_AV" && <iframe src={'https://t.bilibili.com/' + select.OID}  className={'w-full h-full'}/>}
                    </div>
                        }

                </div>
            </div>
        </div>
    );
}

export default ReactionPage;


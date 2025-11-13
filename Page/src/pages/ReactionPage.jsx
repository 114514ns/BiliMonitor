import React, {useState} from 'react';
import {Button, Input, ToastProvider,Image} from "@heroui/react";
import axios from "axios";





function ReactionPage(props) {

    const [data,setData] = useState([])
    const [uid,setUid] = useState('')
    const [ups,setUps] = useState([])
    const [select,setSelect] = useState({})
    const [filtered,setFiltered] = useState([])
    return (
        <div>
            <ToastProvider/>
            <div className={'flex flex-row items-center'}>
                <Input label={'UID'} className={'max-w-xs'} value={uid} onValueChange={(e) => {
                    setUid(e)
                }}></Input>
                <Button className={'ml-2'} onClick={() => {
                    axios.get("/api/reaction?mid=" + uid).then(res =>{
                        setData(res.data.data)
                        setFiltered(res.data.data)
                        var map = new Map()
                        res.data.data.forEach((item,i)=>{
                            map.set(item.TargetName,item.TargetID)
                        })
                        var dst = []
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
            <div className={'flex lg:flex-row flex-col'}>
                {              <div className={'w-[10vw] mr-4 max-h-[85vh] overflow-scroll '}>
                    {(ups).map((key)=>{
                        return (
                            <div className={'flex flex-row items-center mt-2 hover:bg-[#F3F4F5]'} onClick={(e) => {
                                var t = data.filter((item) => item.TargetID === key.UID)
                                setFiltered(t)
                            }}>
                                <img
                                    src={`${AVATAR_API}${key.UID}`}
                                    className="w-12 h-12"
                                    style={{ borderRadius: '50%' }}
                                />
                                <p
                                    className="text-center text-sm truncate w-full hover:text-[#00a1d6]"
                                    title={key.UName}
                                >{key.UName}</p>
                            </div>
                        )
                    })}
                </div>}
                <div className={'overflow-y-auto max-h-[85vh] overflow-x-hidden'}>
                    {filtered.map((item,i)=>{
                        return (
                            <DynamicCard item={item}/>
                        )
                    })}
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


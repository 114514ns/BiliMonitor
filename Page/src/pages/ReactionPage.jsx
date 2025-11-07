import React, {useState} from 'react';
import {Button, Input, ToastProvider,Image} from "@heroui/react";
import axios from "axios";
import { Heart, MessageCircle, Share2, MoreVertical } from 'lucide-react';


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
                            <div className="max-w-2xl mx-auto bg-white rounded-lg shadow-sm p-4 my-8" onClick={() => {
                                setSelect(item)
                            }}>
                                {/* Header */}
                                <div className="flex items-start justify-between mb-3">
                                    <div className="flex items-center gap-3">
                                        <div className="w-12 h-12 bg-purple-200 rounded-full flex items-center justify-center overflow-hidden" >
                                            <img
                                                src={`${AVATAR_API}${item.TargetID}`}
                                                className="w-12 h-12"
                                                style={{ borderRadius: '50%' }}
                                            />
                                        </div>
                                        <div>
                                            <h3 className="font-medium text-gray-900">{item.TargetName}</h3>
                                            <p className="text-sm text-gray-500">{formatTimeAgo(item.CreatedAt)}</p>
                                        </div>
                                    </div>
                                </div>

                                <div className="mb-3">
                                    <p className="text-gray-800 leading-relaxed">
                                        {item.Text}
                                    </p>
                                </div>

                                <div className="mb-4">
                                    {item.Images.split(',').map((item,i)=>{
                                        return (
                                            <Image src={item} className={'pt-4'}/>
                                        )
                                    })}

                                </div>
                                {item.Title}
                                {/* Action Bar */}
                                <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                                    <button className="flex items-center gap-2 text-gray-600 hover:text-gray-800 transition-colors">
                                        <Share2 size={20} />
                                        <span className="text-sm">转发</span>
                                    </button>

                                    <button className="flex items-center gap-2 text-gray-600 hover:text-gray-800 transition-colors">
                                        <MessageCircle size={20} />
                                        <span className="text-sm">{item.Comments}</span>
                                    </button>

                                    <button
                                        className={`flex items-center gap-2 transition-colors ${
                                            'text-gray-600 hover:text-red-500'
                                        }`}
                                    >
                                        <Heart
                                            size={20}
                                            fill={ 'none'}
                                            className="transition-all"
                                        />
                                        <span className="text-sm">{item.Like}</span>
                                    </button>
                                </div>
                            </div>
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
function formatTimeAgo(timeStr) {
    // 支持 "2025-05-15T04:05:02Z" 和 "2025-05-15 12:00:00"
    const date = timeStr.includes('T')
        ? new Date(timeStr)
        : new Date(timeStr.replace(/-/g, '/'))

    const now = new Date()
    const diffMs = now - date
    const diffSec = Math.floor(diffMs / 1000)
    const diffMin = Math.floor(diffSec / 60)
    const diffHour = Math.floor(diffMin / 60)
    const diffDay = Math.floor(diffHour / 24)

    if (diffSec < 60) return '刚刚'
    if (diffMin < 60) return `${diffMin} 分钟前`
    if (diffHour < 24 && now.getDate() === date.getDate()) return `${diffHour} 小时前`

    // 昨天
    const yesterday = new Date()
    yesterday.setDate(now.getDate() - 1)
    if (
        date.getDate() === yesterday.getDate() &&
        date.getMonth() === yesterday.getMonth() &&
        date.getFullYear() === yesterday.getFullYear()
    ) {
        return `昨天 ${date.getHours().toString().padStart(2, '0')}:${date
            .getMinutes()
            .toString()
            .padStart(2, '0')}`
    }

    // 是否同一年
    const sameYear = date.getFullYear() === now.getFullYear()
    const month = date.getMonth() + 1
    const day = date.getDate()

    if (!sameYear) {
        return `${date.getFullYear()}年${month}月${day}日`
    }

    return `${month}月${day}日`
}
export default ReactionPage;
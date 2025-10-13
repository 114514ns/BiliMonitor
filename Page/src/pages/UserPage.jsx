import React, { useEffect, useState } from 'react';
import { useParams } from "react-router-dom";
import axios from "axios";
import {
    Autocomplete, AutocompleteItem, Avatar,
    Tooltip, Button, Checkbox
} from "@heroui/react";
import ActionTable from "../components/ActionTable";
import HoverMedals from "../components/HoverMedals";
import { HeroUIPieChart } from "../components/PieChart";


function UserPage(props) {
    let { id } = useParams();
    const [space, setSpace] = useState({})
    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/user/space?uid=${id}`).then((response) => {
            setSpace(response.data);
            console.log(response.data);
        })
    }, [])
    const [total, setTotal] = useState(1)
    const [liver, setLiver] = useState(0)
    const [data, setData] = useState([])

    const [filter, setFilter] = useState("")

    const [order, setOrder] = useState("")

    const [room, setRoom] = useState("")

    const [input, setInput] = useState("")

    const [showMedal, setShowMedal] = useState(false)

    const [showEnter, setShowEnter] = useState(false)
    let page = 1

    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page}&order=${order}&type=${filter}&room=${room === null ? "" : room} &enter=${showEnter ? '1' : ''} `).then((response) => {
            setData(response.data.data);
            setTotal(response.data.total);
        })
    }, [order, filter, room, showEnter])

    return (
        <div>
            <div className={'flex flex-col sm:flex-row  h-full'}>
                <HeroUIPieChart
                    width={isMobile() ? vwToPx(90) : vwToPx(35)}
                    data={getPieData(space.Rooms)}
                    onSegmentClick={(data, index) => {
                        console.log(data); // 包含所有字段
                        setRoom(data.payload.id)
                    }}
                />
                <div className={'sm:w-[75vw]'}>
                    <div className="grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm ">
                        <div
                            className=" bg-blue-100 p-2 rounded-xl transition-transform transform duration-200 hover:scale-105 hover:shadow-lg cursor-pointer ">
                            <span className="text-blue-600"></span>
                            <div className='flex flex-row items-center text-blue-600' onClick={() => {
                                toSpace(id)
                            }}>
                                <img
                                    src={`${AVATAR_API}${id}`}
                                    className='w-12 h-12 ml-4 mr-4 ' style={{ borderRadius: '50%' }}></img>
                                {space.UName}
                            </div>

                        </div>
                        <div
                            className="rounded-xl bg-gray-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">首次出现<br />
                            <span
                                className="font-semibold">{new Date(space.FirstSeen).toLocaleString()}</span>
                        </div>
                        <div
                            className="rounded-xl bg-pink-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">最后出现<br />
                            <span
                                className="font-semibold">{new Date(space.LastSeen).toLocaleString()}</span>
                        </div>
                    </div>
                    <div className={'grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4'}>
                        <div
                            className="rounded-xl bg-green-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">弹幕<br />
                            <span
                                className="font-semibold">{space.Message}</span>
                        </div>
                        <div
                            className="rounded-xl bg-orange-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">消费<br />
                            <span
                                className="font-semibold">{space.Money}</span>
                        </div>
                        <Tooltip content={<HoverMedals mid={id} />} isOpen={showMedal} onOpenChange={(o) => {
                            setShowMedal(o)
                        }}>
                            <div
                                className="rounded-xl bg-red-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg " onContextMenu={() => {
                                    console.log("context menu")
                                    setShowMedal(true)
                                }} >最高粉丝牌等级<br />
                                <span
                                    className="font-semibold">{space.HighestLevel}</span>
                            </div>
                        </Tooltip>

                    </div>

                    <div className={'mt-4'}>
                        <div className='flex  items-center flex-col sm:flex-row'>
                            <Autocomplete
                                isClearable
                                onClear={() => setFilter('')}
                                className="w-full sm:max-w-xs mt-4 mb-4"
                                defaultItems={[{
                                    key: 'msg',
                                    value: "Message"
                                },
                                {
                                    key: 'gift',
                                    value: "Gift"

                                },
                                {
                                    key: 'guard',
                                    value: "Membership"
                                },
                                {
                                    key: 'sc',
                                    value: "SuperChat"
                                }
                                ]}
                                label="Filter by"
                                onSelectionChange={e => {
                                    setFilter(e)
                                }}
                            >
                                {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
                            </Autocomplete>
                            <Autocomplete
                                isClearable
                                setOrder={() => setFilter('')}
                                className="mt-4 mb-4 sm:ml-4 w-full sm:max-w-xs"
                                defaultItems={[{
                                    key: 'money',
                                    value: "Money"
                                },
                                {
                                    key: 'timeDesc',
                                    value: "Time Desc"

                                },
                                ]}
                                label="Sort by"
                                onSelectionChange={e => {
                                    setOrder(e)
                                }}
                            >
                                {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
                            </Autocomplete>
                            <Autocomplete
                                isClearable
                                setOrder={() => setRoom('')}
                                className=" mt-4 mb-4 sm:ml-4 w-full sm:max-w-xs"
                                label="Liver"
                                onSelectionChange={e => {
                                    setRoom(e)
                                }}
                                onInputChange={e => {
                                    setInput(e)
                                }}
                                selectedKey={room}
                                items={space.Rooms == null ? [] : space.Rooms.sort((a, b) => { return a.Rate - b.Rate }).filter(e => { return e.Liver.includes(input) !== 0 })}
                            >
                                {(f) => <AutocompleteItem key={f.LiveRoom} textValue={f.Liver}>
                                    <div className={'flex flex-row'}>
                                        <Avatar src={`${AVATAR_API}${f.LiverID}`} />
                                        <span className={'font-bold ml-2 mt-2'}>{f.Liver}</span>
                                    </div>
                                </AutocompleteItem>}
                            </Autocomplete>
                            <Checkbox isSelected={showEnter} onValueChange={(e) => { setShowEnter(e) }} className='ml-2 self-start sm:self-auto mb-1 sm:mb-0'>Enter</Checkbox>
                        </div>
                        <ActionTable dataSource={data} handlePageChange={(page0, pageSize) => {
                            page = page0
                            axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page0}&order=${order}&type=${filter}&room=${room === null ? "" : room}&enter=${showEnter ? '1' : ''} `).then((response) => {
                                setData(response.data.data);
                                setTotal(response.data.total);
                            })
                            if (page >= 2) {
                                window.USER_PAGE = page0
                            }
                        }} total={total} />
                    </div>
                </div>
            </div>
        </div>
    );
}


function getPieData(obj) {
    var data = [];
    if (obj instanceof Array) {
        obj.forEach(item => {
            data.push({
                id: item.LiveRoom,
                value: item.Rate * 100,
                name: item.Liver,
            })
        })
    }


    return data
}


export default UserPage;
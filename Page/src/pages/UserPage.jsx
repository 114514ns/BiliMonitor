import React, {useEffect, useMemo, useState} from 'react';
import { useParams } from "react-router-dom";
import axios from "axios";
import {
    Autocomplete, AutocompleteItem, Avatar,
    Tooltip, Button, Checkbox, Modal, useDisclosure, ModalContent, ModalBody, ModalHeader, ModalFooter, Select,
    SelectItem, addToast
} from "@heroui/react";
import ActionTable from "../components/ActionTable";
import HoverMedals from "../components/HoverMedals";
import { HeroUIPieChart } from "../components/PieChart";
import {HeatContent} from "../components/HeatChart";
import {useNavigate} from "react-router";
import MysteryBoxStatistic from "../components/MysteryBoxStatistic";


function UserPage(props) {
    let { id } = useParams();
    const [space, setSpace] = useState({})
    const redirect = useNavigate()
    useEffect(() => {
        const nums = id.match(/\d+/g)?.[0];
        if (nums !== id) {
            redirect('/user/' + nums)
        } else {
            axios.get(`${protocol}://${host}:${port}/api/user/space?uid=${id}`).then((response) => {
                setSpace(response.data);
                document.title = response.data.UName + ' 的弹幕数据'
                console.log(response.data);
            })
        }


        return(() => {
            document.title = 'Vtuber 数据'
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

    const [showBox,setShowBox] = useState(false)
    let page = 1

    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page}&pageSize=${localStorage.getItem("defaultPageSize")}&order=${order}&type=${filter}&room=${room === null ? "" : room} &enter=${showEnter ? '1' : ''} `).then((response) => {
            setData(response.data.data);
            setTotal(response.data.total);

            if (response.data.total === 0) {
                if (filter === '' ) {
                    setNoFound(response.data.total === 0)
                    addToast({
                        title:'没有数据喵',
                        color:'danger'
                    })
                }
            }
        })
    }, [order, filter, room, showEnter])

    const tableRef = React.createRef();
    const {isOpen, onOpen, onOpenChange} = useDisclosure();

    const chartWidth = useMemo(() => {
        return isMobile() ? vwToPx(90) : vhToPx(80);
    }, []); // 窗口尺寸在组件生命周期内一般不变，空依赖即可

    const [noFound,setNoFound] = useState(false)


    return (
        <div>
            <MysteryBoxStatistic isOpen={showBox} onClose={() => {
                setShowBox(false)
            }} type={'user'} uid={id}/>
            <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="2xl">
                <ModalContent>
                    {(onClose) => (
                        <>
                            <ModalHeader className="flex flex-col gap-1">Modal Title</ModalHeader>
                            <ModalBody>
                                <HeatContent uid={id}/>
                            </ModalBody>
                            <ModalFooter>
                                <Button color="danger" variant="light" onPress={onClose}>
                                    Close
                                </Button>
                                <Button color="primary" onPress={onClose}>
                                    Action
                                </Button>
                            </ModalFooter>
                        </>
                    )}
                </ModalContent>
            </Modal>
            {!noFound &&             <div className={'flex flex-col sm:flex-row  h-[88vh] '}>
                <HeroUIPieChart
                    width={chartWidth}
                    data={getPieData(space.Rooms)}
                    onSegmentClick={(data, index) => {
                        console.log(data); // 包含所有字段
                        setRoom(data.payload.id + '')
                    }}
                />
                <div className={'sm:w-[75vw] lg:overflow-x-hidden '}>
                    <div className="grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm ">
                        <div
                            className=" bg-blue-100 dark:bg-gray-500 p-2 rounded-xl transition-transform transform duration-200 hover:scale-105 hover:shadow-lg cursor-pointer ">
                            <span className="text-blue-600"></span>
                            <div className='flex flex-row items-center text-blue-600 dark:text-stone-100' onClick={() => {
                                toSpace(id)
                            }}>
                                <img
                                    src={`${AVATAR_API}${id}`}
                                    className='w-12 h-12 ml-4 mr-4 ' style={{ borderRadius: '50%' }}></img>
                                {space.UName}
                            </div>

                        </div>
                        <div
                            className="rounded-xl bg-gray-100 dark:bg-gray-500 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg " onClick={onOpen}>首次出现<br />
                            <span
                                className="font-semibold">{new Date(space.FirstSeen).toLocaleString()}</span>
                        </div>
                        <div
                            className="rounded-xl bg-pink-100 dark:bg-gray-500  p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">最后出现<br />
                            <span
                                className="font-semibold">{new Date(space.LastSeen).toLocaleString()}</span>
                        </div>
                    </div>
                    <div className={'grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4'}>
                        <div
                            className="rounded-xl bg-green-100 dark:bg-gray-500 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">弹幕<br />
                            <span
                                className="font-semibold">{space.Message}</span>
                        </div>
                        <Tooltip     content={<div>
                            <p>大航海：{space.GuardMoney && space.GuardMoney.toLocaleString()}</p>
                            <p>礼物/SC：{space.GiftMoney && space.GiftMoney.toLocaleString()}</p>
                        </div>}>
                            <div
                                onClick={() => {
                                    setShowBox(true)
                                }}
                                className="rounded-xl bg-orange-100 dark:bg-gray-500 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">消费<br />

                                <span
                                    className="font-semibold">{(space.GiftMoney + space.GuardMoney).toLocaleString()}</span>

                            </div>

                        </Tooltip>
                        <Tooltip content={<HoverMedals mid={id} />} isOpen={showMedal} onOpenChange={(o) => {
                            setShowMedal(o)
                        }}>
                            <div
                                className="rounded-xl bg-red-100 dark:bg-gray-500 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg " onContextMenu={() => {
                                console.log("context menu")
                                setShowMedal(true)
                            }} >最高粉丝牌等级<br />
                                <span
                                    className="font-semibold">{space.HighestLevel}</span>
                            </div>
                        </Tooltip>

                    </div>

                    <div className={'mt-4 ' } ref={tableRef}>
                        <div className='flex  items-center flex-col sm:flex-row '>
                            <Select
                                isClearable
                                onClear={() => setFilter('')}
                                className="w-full sm:max-w-xs mt-4 mb-4"
                                items={[{
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
                                    },
                                    {
                                        key: 'box',
                                        value: "MysteryBox"
                                    },
                                ]}
                                label="Filter by"
                                onSelectionChange={e => {
                                    setFilter(e.currentKey)
                                }}
                            >
                                {(f) => <SelectItem key={f.key}>{f.value}</SelectItem>}
                            </Select>
                            <Select
                                isClearable
                                setOrder={() => setFilter('')}
                                className="mt-4 mb-4 sm:ml-4 w-full sm:max-w-xs"
                                items={[{
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
                                {(f) => <SelectItem key={f.key}>{f.value}</SelectItem>}
                            </Select>
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
                                {(f) => <AutocompleteItem key={f.LiveRoom + ''} textValue={f.Liver}>
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
                            axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page0}&pageSize=${localStorage.getItem("defaultPageSize")}&order=${order}&type=${filter}&room=${room === null ? "" : room}&enter=${showEnter ? '1' : ''} `).then((response) => {
                                setData(response.data.data);
                                setTotal(response.data.total);
                            })
                            if (page >= 2) {
                                window.USER_PAGE = page0
                            }
                            tableRef.current.parentElement.scroll({
                                top: 0,
                                behavior: "smooth"
                            })
                        }} total={total} />
                    </div>
                </div>
            </div>}
            {noFound && <div className={'flex items-center flex-col'}>
                <img src={'https://i0.hdslb.com/bfs/new_dyn/cabcfbb4745084bec654860c582e4f491995486878.png'}/>
            </div>}
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
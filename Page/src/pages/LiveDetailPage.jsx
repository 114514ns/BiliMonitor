import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import axios from "axios";
import "./LivePage.css"

import {
    Autocomplete,
    AutocompleteItem,
    Avatar, Chip,
    Pagination,
    Table,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow, Tooltip
} from "@heroui/react";
import UserChip from "../components/UserChip";
import {CheckIcon} from "./ChatPage";

function LiveDetailPage(props) {
    let {id} = useParams();
    const host = location.hostname;
    const [actions, setActions] = useState([])

    const [mid,setMid] = useState(0);
    useEffect(() => {
        refreshData(currentPage, pageSize)
    }, [])
    const [currentPage, setCurrentPage] = useState(1);

    const [pageSize, setPageSize] = useState(10);
    const [dataSource, setDatasource] = useState([])

    const [total, setTotal] = useState(0)

    const [selected, isSelected] = useState(false)

    const [name, setName] = useState(null)
    const [order, setOrder] = useState("undefined")
    const [filters, setFilters] = useState([
        {text: 'Joe', value: 'Joe'},
        {text: 'Jim', value: 'Jim'},
        {text: 'Category 1', value: 'Category 1'},
        {text: 'Category 2', value: 'Category 2'},
    ]);
    const [columns, setColumn] = useState([])

    const [liveInfo, setLiveInfo] = useState({});


    const [user, setUser] = useState([]);
    useEffect(() => {
        refreshData(currentPage, pageSize)
    }, [order])

    const [filter, setFilter] = useState('')

    useEffect(() => {

        setColumn([
            {
                title: 'Name',
                dataIndex: 'FromName',
                key: 'UserName',
                filterSearch: true,
                filters: filters,
                render: (text, record) => (
                    <span style={{cursor: 'pointer'}} onClick={() => {
                        window.open("https://space.bilibili.com/" + record.FromId)
                    }}>
        {text}{console.log(record)}
      </span>
                )
            },
            {
                title: 'Title',
                dataIndex: 'Liver',
                key: 'Title',
            },
            {
                title: 'Time',
                dataIndex: 'CreatedAt',
                key: 'StartAt',
            },
            {
                title: 'Money',
                dataIndex: 'GiftPrice',
                key: 'Money',
                sorter: true,

            },
            {
                title: 'Message',
                dataIndex: 'Extra',
                key: 'Message'
            }
        ])
    }, [])
    const port = location.port
    const protocol = location.protocol.replace(":", "")
    const refreshData = (page, size, name) => {
        if (page === undefined) {
            return
        }
        var url = `${protocol}://${host}:${port}/api/live/` + id + "/?" + "page=" + page + "&limit=" + size + "&order=" + order + "&mid=" + mid
        if (name != null) {
            url = url + `&name=${name}`
        }
        if (filter != null) {
            url = url + `&type=${filter}`
        }
        axios.get(url).then(res => {

            res.data.records.forEach((item, index) => {
                if (item.GiftName != "") {
                    res.data.records[index].Extra = item.GiftName
                }
                res.data.records[index].Liver = res.data.liver
                res.data.records[index].GiftPrice = res.data.records[index].GiftPrice.Float64
                res.data.records[index].CreatedAt = new Date(res.data.records[index].CreatedAt).toLocaleString()
            })
            setTotal(res.data.totalPages * size)
            console.log(total)
            setDatasource(res.data.records)
        })
    }
    useEffect(() => {
        var url = `${protocol}://${host}:${port}/api/liveDetail/` + id + "/"
        axios.get(url).then(res => {
            setLiveInfo(res.data.live)
        })
        axios.get(`${protocol}://${host}:${port}/api/liveUser?live=${id}`).then(res => {
            setUser(res.data.list)
        })
    }, []);

    useEffect(() => {
        refreshData(currentPage, pageSize)
    },[filter,mid])


    const handlePageChange = (page, pageSize, sorter) => {
        refreshData(page, pageSize, name)
        setCurrentPage(page)
        setPageSize(pageSize)
        console.log(sorter)

    }

    function formatTimeDiff(startTimestamp, endTimestamp) {
        const diffMs = Math.abs(endTimestamp - startTimestamp); // 毫秒差
        const diffMinutes = Math.floor(diffMs / 1000 / 60);

        if (diffMinutes <= 60) {
            return `${diffMinutes} 分钟`;
        } else {
            const diffHours = (diffMinutes / 60).toFixed(1); // 保留 1 位小数
            return `${diffHours} 小时`;
        }
    }


    return (
        <div>
            <div className="flex  space-x-4 rounded-2xl bg-white p-4 shadow-md">
                <div className="flex-1 space-y-2">
                    <h2 className="text-xl font-bold">{liveInfo.Title}</h2>
                    <div className="grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm ">
                        <div
                            className=" bg-blue-100 p-2 rounded-xl transition-transform transform duration-200 hover:scale-105 hover:shadow-lg cursor-pointer ">
                            <span className="text-blue-600">主播</span>
                            <div className='flex flex-row items-center' onClick={() => {toSpace(liveInfo.UserID)}}>
                                <img src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${liveInfo.UserID}`} className='w-12 h-12 ml-4 mr-4 ' style={{borderRadius:'50%'}}></img>
                                <br/>
                                {liveInfo.UserName}
                            </div>

                        </div>
                        <div
                            className="rounded-xl bg-gray-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">房间号<br/>
                            <span
                            className="font-semibold">{liveInfo.RoomId}</span>
                        </div>
                        <div
                            className="rounded-xl bg-gray-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">分区<br/><span className="font-semibold">{liveInfo.Area}</span>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                        <div
                            className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">开始时间<br/><span
                            className="font-semibold">{new Date(liveInfo.StartAt * 1000-8*3600*1000).toLocaleString()}</span>
                        </div>
                        <div
                            className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">结束时间<br/>
                            <span className="font-semibold">{new Date(liveInfo.EndAt * 1000).toLocaleString()}</span>
                        </div>
                        <div
                            className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">时长<br/>
                            <span
                                className="font-semibold">{formatTimeDiff(liveInfo.StartAt * 1000-8*3600*1000, liveInfo.EndAt * 1000)}</span>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3  gap-2 text-sm">
                        <div className="rounded-xl bg-green-100 p-2 text-green-700 transition-transform duration-200 hover:scale-105 hover:shadow-lg">观众数<br/>{liveInfo.Watch}</div>
                        <div className="rounded-xl bg-purple-100 p-2 text-fuchsia-600 transition-transform duration-200 hover:scale-105 hover:shadow-lg">弹幕数<br/>{liveInfo.Message}
                        </div>
                        <div className="rounded-xl bg-rose-100 p-2 text-rose-600 transition-transform duration-200 hover:scale-105 hover:shadow-lg">流水<br/>{liveInfo.Money}</div>
                    </div>
                </div>
            </div>

            <Autocomplete
                className="max-w-xs mt-4 mb-4 ml-4"
                defaultItems={[{
                    key: 'ascend',
                    value: "Ascend"
                },
                    {
                        key: 'descend',
                        value: "Descend"

                    },
                    {
                        key: 'Time',
                        value: "Time"
                    }
                ]}
                label="Sort by"
                onSelectionChange={e => {
                    setOrder(e)
                }}
            >
                {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
            </Autocomplete>
            <Autocomplete
                className="max-w-xs mt-4 mb-4 ml-4"
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
                className="max-w-xs mt-4 mb-4 ml-4"
                items={user}
                label="Search Watcher"
                onSelectionChange={e => {
                    setMid(e)
                }}
                onInputChange={e => {
                    var url = `${protocol}://${host}:${port}/api/liveUser?live=${id}&keyword=${e}`;
                    axios.get(url).then((response) => {
                        setUser(response.data.list)
                    })
                }}
            >
                {(f) => <AutocompleteItem key={f.FromId} textValue={f.FromName}>
                    <UserChip props={f}></UserChip>
                </AutocompleteItem>}
            </Autocomplete>
            <Table bottomContent={
                <div className="flex w-full justify-center">
                    <Pagination
                        isCompact
                        showControls
                        showShadow
                        color="secondary"
                        page={currentPage}
                        total={total / pageSize}
                        onChange={(page) => handlePageChange(page, pageSize)}
                    />
                </div>
            } isStriped>

                <TableHeader>
                    {columns.map((col, index) => (
                        <TableColumn key={index}>{col.title}</TableColumn>

                    ))}
                </TableHeader>
                <TableBody>

                    {dataSource.map((item, index) => (
                        <TableRow key={index} onClick={() => {

                        }}>
                            <TableCell>
                                <Tooltip content={
                                    <HoverMedals mid={item.FromId}/>
                                } delay={400}>
                                    <div className={'flex'} onClick={() => {
                                        toSpace(item.FromId)
                                    }}>
                                        {item.FromName}
                                        {item.MedalLevel != 0 ?                                     <Chip
                                            className={'basis-64'}
                                            startContent={<CheckIcon size={18}/>}
                                            variant="faded"
                                            onClick={() => {
                                                toSpace(props.props.LiverID);
                                            }}
                                            style={{background: getColor(item.MedalLevel), color: 'white', marginLeft: '8px'}}
                                        >
                                            {item.MedalName}
                                            <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {item.MedalLevel}
                                                        </span>
                                        </Chip>:<></>}
                                    </div>
                                </Tooltip>
                            </TableCell>
                            <TableCell>{item.Liver}</TableCell>
                            <TableCell>{item.CreatedAt}</TableCell>
                            <TableCell>{item.GiftPrice}</TableCell>
                            <TableCell>{item.Extra}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

function HoverMedals(props) {
    var mid = props.mid
    var [data,setData] = useState([])
    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/medals?mid=${mid}`).then((response) => {
            if (response.data.list == null) {
                response.data.list = []
            }
            response.data.list.sort((a,b) => {
                return a.Score < b.Score
            })
            setData(response.data.list)
        })
    },[])
    return (

        <div className={'flex flex-col scroll-auto'}>
            {data.map((item, index) => (
                <div key={index}>
                    <p className={'font-medium'}>{item.Liver}</p>

                    <div className={'flex flex-row align-middle mt-2'}>
                        <Avatar
                            src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${item.LiverID}`}
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}/>

                        {item.Level ?              <Chip
                            startContent={item.Type?<img  src={getGuardIcon(item.Type)} className={'w-6 h-6'}/>:<div/>}
                            variant="faded"
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}
                            style={{background: getColor(item.Level), color: 'white', marginLeft: '8px'}}
                        >
                            {item.MedalName}
                            <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {item.Level}
                                                        </span>
                        </Chip>:<></>}
                    </div>
                </div>
            ))}

        </div>
    )
}

function getGuardIcon(level) {
    var array = ["","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-1.b718085..png","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-2.d43d078..png","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-3.6d2f428..png"]
    return array[level]
}

export default LiveDetailPage;
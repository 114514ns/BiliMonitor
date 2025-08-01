import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import axios from "axios";
import {PieChart} from '@mui/x-charts/PieChart';
import {
    Autocomplete, AutocompleteItem,
    Chip,
    Pagination,
    Table,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow,
    Tooltip
} from "@heroui/react";
import {CheckIcon} from "./ChatPage";
import ActionTable from "../components/ActionTable";

function isMobile() {
    return /Mobi|Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)  || window.innerWidth <= 768
}
function UserPage(props) {
    let {id} = useParams();
    const [space, setSpace] = useState({})
    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/user/space?uid=${id}`).then((response) => {
            setSpace(response.data);
            console.log(response.data);
        })
    }, [])
    const [total, setTotal] = useState(1)
    const [liver,setLiver] = useState(0)
    const [data,setData] = useState([])

    const [filter, setFilter] = useState("")

    const [order, setOrder] = useState("")

    let page = 1

    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page}&order=${order}&type=${filter}`).then((response) => {
            setData(response.data.data);
            setTotal(response.data.total);
        })
    },[order,filter])

    return (
        <div>
            <div className={'flex flex-col sm:flex-row  h-full'}>
                <PieChart
                    series={[
                        {
                            data: getPieData(space.Rooms)
                        },
                    ]}
                    width={isMobile()?vwToPx(90):vwToPx(35)}
                    height={vhToPx(35)}
                    onItemClick={(event,item) => {
                        console.log(space.Rooms[item.dataIndex]);
                    }}
                />
                <div className={'sm:w-[65vw]' }>
                    <div className="grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm ">
                        <div
                            className=" bg-blue-100 p-2 rounded-xl transition-transform transform duration-200 hover:scale-105 hover:shadow-lg cursor-pointer ">
                            <span className="text-blue-600"></span>
                            <div className='flex flex-row items-center text-blue-600' onClick={() => {
                                toSpace(liveInfo.UserID)
                            }}>
                                <img
                                    src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${id}`}
                                    className='w-12 h-12 ml-4 mr-4 ' style={{borderRadius: '50%'}}></img>
                                <br/>
                                {space.UName}
                            </div>

                        </div>
                        <div
                            className="rounded-xl bg-gray-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">首次出现<br/>
                            <span
                                className="font-semibold">{new Date(space.FirstSeen).toLocaleString()}</span>
                        </div>
                        <div
                            className="rounded-xl bg-pink-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">最后出现<br/>
                            <span
                                className="font-semibold">{new Date(space.LastSeen).toLocaleString()}</span>
                        </div>
                    </div>
                    <div className={'grid  grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4'}>
                        <div
                            className="rounded-xl bg-green-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">弹幕<br/>
                            <span
                                className="font-semibold">{space.Message}</span>
                        </div>
                        <div
                            className="rounded-xl bg-orange-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">消费<br/>
                            <span
                                className="font-semibold">{space.Money}</span>
                        </div>
                        <div
                            className="rounded-xl bg-red-100 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">最高粉丝牌等级<br/>
                            <span
                                className="font-semibold">{space.HighestLevel}</span>
                        </div>
                    </div>

                    <div className={'mt-4'}>
                        <div>
                            <Autocomplete
                                className="max-w-xs mt-4 mb-4"
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
                                className="max-w-xs mt-4 mb-4 sm:ml-4 sm:w-full"
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
                        </div>
                        <ActionTable dataSource={data} handlePageChange={(page0,pageSize) => {
                            page = page0
                            axios.get(`${protocol}://${host}:${port}/api/user/action?uid=${id}&page=${page0}&order=${order}&type=${filter}`).then((response) => {
                                setData(response.data.data);
                                setTotal(response.data.total);
                            })
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
                label: item.Liver,
            })
        })
    }


    return data
}

function vhToPx(vhPercent) {
    const vh = window.innerHeight;
    return (vhPercent / 100) * vh;
}

function vwToPx(vhPercent) {
    const vh = window.innerWidth;
    return (vhPercent / 100) * vh;
}

export default UserPage;
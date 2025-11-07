import React, {useEffect} from 'react';
import {addToast, Autocomplete, AutocompleteItem, Avatar, Select, SelectItem, ToastProvider} from "@heroui/react";
import ActionTable from "../components/ActionTable";
import axios from "axios";

function RawPage(props) {
    const [filter, setFilter] = React.useState("");

    const [room, setRoom] = React.useState("");

    const [input, setInput] = React.useState("");

    const [roomList, setRoomList] = React.useState([]);

    const [order, setOrder] = React.useState("");

    const [data,setData] = React.useState([]);

    const [total,setTotal] = React.useState(0);

    const [pageSize, setPageSize] = React.useState(10);
    useEffect(() => {
        axios.get("/api/searchLiver?key=" + input).then((response) => {
            setRoomList(response.data.result??[])
        })
    },[input])
    useEffect(() => {
        console.log(roomList);
    }, [roomList]);
    useEffect(() => {
        axios.get(`/api/raw?room=${room}&type=${filter}&page=${page}&order=${order}&size=${pageSize}`).then((response) => {
            setData(response.data.data);
            setTotal(response.data.total);
        }).catch((err,res) => {
            addToast({
                title: "Error",
                description: err.data.message,
                hideIcon: true,
            });
        })
    },[order,filter,room,pageSize])
    return (<div>
        <ToastProvider/>
            <div className={'mt-4'}>
                <div>
                    <Autocomplete
                        className="w-full sm:max-w-xs mt-4 mb-4"
                        defaultItems={[{
                            key: '1', value: "Message"
                        }, {
                            key: '2', value: "Gift"

                        }, {
                            key: '3', value: "Membership"
                        }, {
                            key: '4', value: "SuperChat"
                        }]}
                        label="Filter by"
                        onSelectionChange={e => {
                            setFilter(e)
                        }}
                    >
                        {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
                    </Autocomplete>
                    <Autocomplete
                        className="mt-4 mb-4 sm:ml-4 w-full sm:max-w-xs"
                        defaultItems={[{
                            key: 'money_desc', value: "Money"
                        }, {
                            key: 'created_at_desc', value: "Time Desc"

                        },]}
                        label="Sort by"
                        onSelectionChange={e => {
                            setOrder(e)
                        }}
                    >
                        {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
                    </Autocomplete>
                    <Autocomplete
                        className=" mt-4 mb-4 sm:ml-4 w-full sm:max-w-xs"
                        label="Liver"
                        onSelectionChange={e => {
                            setRoom(e)
                        }}
                        onInputChange={e => {
                            setInput(e)
                        }}
                        items={roomList}
                    >
                        {(f) => <AutocompleteItem key={f.Room} textValue={f.UName} className={''}>
                            <div className={'flex flex-row '}>
                                <Avatar src={`${AVATAR_API}${f.UID}`}/>
                                <span className={'font-bold ml-2 mt-2'}>{f.UName}</span>
                            </div>
                        </AutocompleteItem>}
                    </Autocomplete>
                    <Select className="max-w-xs mt-4 mb-4 ml-4" label={'Page Size'} defaultSelectedKeys={['10']}>
                        <SelectItem onClick={e => { setPageSize(10) }} key={'10'}>
                            10
                        </SelectItem>
                        <SelectItem onClick={e => { setPageSize(50)}}>
                            50
                        </SelectItem>
                        <SelectItem onClick={e => { setPageSize(200)}}>
                            200
                        </SelectItem>
                        <SelectItem onClick={e => { setPageSize(500)}}>
                            500
                        </SelectItem>
                    </Select>
                </div>
                <ActionTable dataSource={data} handlePageChange={(page0, pageSiz) => {
                    page = page0
                    axios.get(`/api/raw?room=${room}&type=${filter}&page=${page0}&order=${order}&size=${pageSize}`).then((response) => {
                        setData(response.data.data);
                        setTotal(response.data.total);
                    })
                    if (page >= 2) {
                        window.USER_PAGE = page0
                    }
                }} total={total}/>
            </div>
        </div>);
}

export default RawPage;
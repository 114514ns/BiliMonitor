import React, {useEffect, useState} from 'react';
import {FloatButton, Input} from "antd";
import axios from "axios";
import {useNavigate} from "react-router";
import "./LivePage.css"
import {
    Autocomplete, AutocompleteItem,
    Pagination,
    Table,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow,
} from "@heroui/react";

function LivePage(props) {


    const [dataSource, setDatasource] = useState([])

    const [total, setTotal] = useState(0)

    const [selected, isSelected] = useState(false)

    const [name, setName] = useState(null)

    const [searchText, setSearchText] = useState("");

    const host = location.hostname;

    const port = location.port

    const protocol = location.protocol.replace(":", "")

    const redirect = useNavigate()
    const refreshData = (page, size, name) => {
        var url = `${protocol}://${host}:${port}/live?page=` + page + "&limit=" + size
        if (name != null) {
            url = url + `&name=${name}`
        }
        axios.get(url).then(res => {

            res.data.lives.forEach((item, index) => {
                if (item.EndAt == 0) {
                    res.data.lives[index].EndAt = "直播中"
                } else {
                    res.data.lives[index].EndAt = new Date(item.EndAt * 1000).toLocaleString()
                }
                res.data.lives[index].StartAt = new Date(item.StartAt * 1000 - 8 * 3600 * 1000).toLocaleString()
                //res.data.lives[index].EndAt = new Date(item.EndAt * 1000).toLocaleString()
            })
            setTotal(res.data.totalPage * size)
            console.log(total)
            setDatasource(res.data.lives)
        })
    }

    useEffect(() => {
        refreshData(1, 10)
    }, [])

    const [filters, setFilters] = useState([]);
    useEffect(() => {
        var arr = []
        dataSource.forEach(item => {
            arr.push({
                key:item.UserName,
                value:item.UserName
            })
        })
        setFilters(arr)
    },[dataSource])

    const [columns, setColumn] = useState([
        {
            title: 'Name',
            dataIndex: 'UserName',
            key: 'UserName',
            filterSearch: true,
            filters: filters,

        },
        {
            title: 'Title',
            dataIndex: 'Title',
            key: 'Title',
        },
        {
            title: 'Time',
            dataIndex: 'StartAt',
            key: 'StartAt',
        },
        {
            title: 'EndAt',
            dataIndex: 'EndAt',
            key: 'EndAt'
        },
        {
            title: 'Area',
            dataIndex: 'Area',
            key: 'Area',
        },
        {
            title: 'Money',
            dataIndex: 'Money',
            key: 'Money',
            render: (text) => (
                <span style={{color: text > 1000 ? "red" : "green"}}>
        {text}
      </span>
            ),

        },
        {
            title: 'Message',
            dataIndex: 'Message',
            key: 'Message'
        }
    ])
    const [currentPage, setCurrentPage] = useState(1);

    const [pageSize, setPageSize] = useState(10);

    // 处理页码改变事件
    const handlePageChange = (page, pageSize) => {
        console.log(`page=${page}  pageSize=${pageSize}`)
        refreshData(page, pageSize, name)
        setCurrentPage(page)
        setPageSize(pageSize)

    }

    return (

        <div>
            <FloatButton onClick={() => {
                axios.get(`http://${host}:${port}/refreshMoney`).then(res => {
                    refreshData(currentPage, pageSize)
                })
            }} type="primary">Refresh Money</FloatButton>
            <Autocomplete
                className="max-w-xs"
                defaultItems={filters}
                label="Favorite Animal"
                placeholder="Search an animal"
                onChange={(e) => {
                    axios.get(`${protocol}://${host}:${port}/liver?key=` + e).then(res => {
                        if (!res.data.result) return; // 处理 null/undefined/空数据
                        const newFilters = res.data.result.map((item) => ({ key: item, value: item }));

                        setFilters(newFilters);


                    })
                }}
            >
                {(f) => <AutocompleteItem key={f.key}>{f.value}</AutocompleteItem>}
            </Autocomplete>
            <Table bottomContent={
                <div className="flex w-full justify-center">
                    <Pagination
                        isCompact
                        showControls
                        showShadow
                        color="secondary"
                        page={currentPage}
                        total={total/pageSize}
                        onChange={(page) => handlePageChange(page, pageSize)}
                    />
                </div>
            } rowHeight={70}>

                <TableHeader>
                    {columns.map((col, index) => (
                        <TableColumn key={index}>{col.title}</TableColumn>

                    ))}
                </TableHeader>
                <TableBody>

                    {dataSource.map((item, index) => (
                        <TableRow key={index}>
                            <TableCell>{item.UserName}</TableCell>
                            <TableCell>{item.Title}</TableCell>
                            <TableCell>{item.StartAt}</TableCell>
                            <TableCell>{item.EndAt}</TableCell>
                            <TableCell>{item.Area}</TableCell>
                            <TableCell>{item.Money}</TableCell>
                            <TableCell>{item.Message}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    )
}

export default LivePage
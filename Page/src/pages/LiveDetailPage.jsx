import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import axios from "axios";
import  "./LivePage.css"
import {
    Autocomplete,
    AutocompleteItem,
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownTrigger,
    Pagination,
    Table,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow
} from "@heroui/react";
function LiveDetailPage(props) {
    let { id } = useParams();
    const host = location.hostname;
    const [actions,setActions] = useState([])
    useEffect(() => {
        refreshData(currentPage,pageSize)
    },[])
    const [currentPage, setCurrentPage] = useState(1);

    const [pageSize,setPageSize] = useState(10);
    const [dataSource, setDatasource] = useState([])

    const [total, setTotal] = useState(0)

    const [selected, isSelected] = useState(false)

    const [name,setName] = useState(null)
    const [order, setOrder] = useState("undefined")
    const [filters, setFilters] = useState([
        {text: 'Joe', value: 'Joe'},
        {text: 'Jim', value: 'Jim'},
        {text: 'Category 1', value: 'Category 1'},
        {text: 'Category 2', value: 'Category 2'},
    ]);
    const [columns, setColumn] = useState([])
    useEffect(() => {
        refreshData(currentPage,pageSize)
    },[order])
    useEffect(() => {

        setColumn([
            {
                title: 'Name',
                dataIndex: 'FromName',
                key: 'UserName',
                filterSearch: true,
                filters: filters,
                render: (text,record) => (
                    <span style={{cursor:'pointer'}} onClick={() => {
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
                sorter:true,

            },
            {
                title: 'Message',
                dataIndex: 'Extra',
                key: 'Message'
            }
        ])
    }, [])
    const port = location.port
    const protocol = location.protocol.replace(":","")
    const refreshData = (page, size, name) => {
        if (page === undefined) {
            return
        }
        var url = `${protocol}://${host}:${port}/live/` + id + "/?" +  "page=" + page + "&limit=" + size + "&order=" + order
        if (name != null) {
            url = url + `&name=${name}`
        }
        axios.get(url).then(res => {

                res.data.records.forEach((item, index) => {
                    if (item.GiftName != "") {
                        res.data.records[index].Extra = item.GiftName
                    }
                    res.data.records[index].Liver = res.data.liver
                    res.data.records[index].GiftPrice = res.data.records[index].GiftPrice.Float64
                    res.data.records[index].CreatedAt =  new Date(res.data.records[index].CreatedAt).toLocaleString()
                })
            setTotal(res.data.totalPages * size)
            console.log(total)
            setDatasource(res.data.records)
        })
    }
    // 处理页码改变事件

    const handlePageChange = (page, pageSize,sorter) => {
        refreshData(page, pageSize,name)
        setCurrentPage(page)
        setPageSize(pageSize)
        console.log(sorter)

    }


    return (
        <div>
            <Autocomplete
                className="max-w-xs"
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
            <Table  bottomContent={
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
            }  isStriped>

                <TableHeader>
                    {columns.map((col, index) => (
                        <TableColumn key={index}>{col.title}</TableColumn>

                    ))}
                </TableHeader>
                <TableBody>

                    {dataSource.map((item, index) => (
                        <TableRow key={index} onClick={() => {
                            redirect(`/lives/${record.ID}`)
                        }}>
                            <TableCell>{item.FromName}</TableCell>
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

export default LiveDetailPage;
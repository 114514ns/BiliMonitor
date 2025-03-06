import React, {useEffect, useState} from 'react';
import axios from "axios";
import {useNavigate} from "react-router";
import "./LivePage.css"
import {
    Autocomplete,
    AutocompleteItem, Button,
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

const VerticalDotsIcon = ({size = 24, width, height, ...props}) => {
    return (
        <svg
            aria-hidden="true"
            fill="none"
            focusable="false"
            height={size || height}
            role="presentation"
            viewBox="0 0 24 24"
            width={size || width}
            {...props}
        >
            <path
                d="M12 10c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0-6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 12c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"
                fill="currentColor"
            />
        </svg>
    );
};
const RefreshIcon = (props) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        height="24px"
        viewBox="0 -960 960 960"
        width="24px"
        fill="#1f1f1f"
        {...props}
    >
        <path d="M480-160q-134 0-227-93t-93-227q0-134 93-227t227-93q69 0 132 28.5T720-690v-110h80v280H520v-80h168q-32-56-87.5-88T480-720q-100 0-170 70t-70 170q0 100 70 170t170 70q77 0 139-44t87-116h84q-28 106-114 173t-196 67Z" />
    </svg>
);


function LivePage(props) {


    const [dataSource, setDatasource] = useState([])

    const [total, setTotal] = useState(0)

    const [selected, isSelected] = useState(false)

    const [name, setName] = useState(null)

    const [searchText, setSearchText] = useState("");

    const host = location.hostname;


    const port = debug ? 8080 : location.port

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
                key: item.UserName,
                value: item.UserName
            })
        })
        setFilters(arr)
    }, [dataSource])

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
        },
        {
            title: 'Message',
            dataIndex: 'Message',
            key: 'Message'
        },
        {
            title: 'Action',
            dataIndex: 'Action',
            key: 'Action',
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
            <Button onClick={() => {
                axios.get(`http://${host}:${port}/refreshMoney`).then(res => {
                    refreshData(currentPage, pageSize)
                })
            }} type="primary"  style={{ position: "fixed", bottom: "20px", right: "20px" }}><RefreshIcon/></Button>
            <Autocomplete
                className="max-w-xs"
                defaultItems={filters}
                label="Liver"
                onSelectionChange={e => {
                    refreshData(currentPage, pageSize, e)
                }}
                onInputChange={e => {
                    axios.get(`${protocol}://${host}:${port}/liver?key=` + e).then(res => {
                        if (!res.data.result) return; // 处理 null/undefined/空数据
                        const newFilters = res.data.result.map((item) => ({key: item, value: item}));

                        setFilters(newFilters);
                    })
                }}
                onChange={(e) => {
                    console.log(e)

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
                        total={total / pageSize}
                        onChange={(page) => handlePageChange(page, pageSize)}
                    />
                </div>
            }       maxTableHeight={500}
                   rowHeight={70} isStriped>

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
                            <TableCell>{item.UserName}</TableCell>
                            <TableCell>{item.Title}</TableCell>
                            <TableCell>{item.StartAt}</TableCell>
                            <TableCell>{item.EndAt}</TableCell>
                            <TableCell>{item.Area}</TableCell>
                            <TableCell>{item.Money}</TableCell>
                            <TableCell>{item.Message}</TableCell>
                            <TableCell>
                                <div className="relative flex  items-center gap-2">
                                    <Dropdown>
                                        <DropdownTrigger>
                                            <Button isIconOnly size="sm" variant="light">
                                                <VerticalDotsIcon className="text-default-300"/>
                                            </Button>
                                        </DropdownTrigger>
                                        <DropdownMenu>
                                            <DropdownItem key="view" onClick={() => {
                                                redirect(`/lives/${item.ID}`)
                                            }}>Open</DropdownItem>
                                        </DropdownMenu>
                                    </Dropdown>
                                </div>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    )
}

export default LivePage
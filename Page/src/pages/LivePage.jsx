import React, {useEffect, useState} from 'react';
import axios from "axios";
import {useNavigate} from "react-router";
import "./LivePage.css"
import {
    Autocomplete,
    AutocompleteItem, Button,
    Pagination, Select, SelectItem,
} from "@heroui/react";
import LiveStatisticCard from "../components/LiveStatisticCard";

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

    const [order, setOrder] = useState("id");

    const host = location.hostname;


    const port = location.port

    const protocol = location.protocol.replace(":", "")

    const [chart, setChart] = useState(false)
    const [chartId, setChartId] = useState(0)
    const [liver,setLiver] = useState("")

    const redirect = useNavigate()
    const refreshData = (page, size, name) => {
        var url = `${protocol}://${host}:${port}/api/live?page=` + page + "&limit=" + size
        if (liver != null && liver !== "") {
            url = url + "&uid=" + (liver===-1?"0":liver)
        } else {
            url = url + "&uid=0"
        }
        url = url + "&order=" + order
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
        refreshData(1, pageSize)
    }, [])

    const [filters, setFilters] = useState([]);

    const [currentPage, setCurrentPage] = useState(window.page??1);

    const [pageSize, setPageSize] = useState(20);

    // 处理页码改变事件
    const handlePageChange = (page, pageSize) => {
        console.log(`page=${page}  pageSize=${pageSize}`)
        refreshData(page, pageSize, name)
        setCurrentPage(page)
        setPageSize(pageSize)
        window.page = page;

    }

    useEffect(() => {
        refreshData(currentPage, pageSize)
    },[liver,order])
    useEffect(() => {
        if (chartId !== null && chartId !== 0) {
            setChart(true);
        }
    }, [chartId]); // 监听 chartId 变化后再设置 chart

    return (

        <div className={''}>
            <div className='sm:flex-row flex mb-4 flex-col'>
                <Autocomplete
                    className="max-w-xs mt-4 mb-4 ml-4"
                    items={filters}
                    label="Liver"
                    onSelectionChange={e => {
                        setCurrentPage(1)

                        filters.forEach(filter => {
                            if (filter.UID === parseInt(e)) {
                                setLiver(filter.UID)
                            }
                        })
                        console.log("onSelectionChange")
                    }}
                    onInputChange={e => {
                        axios.get(`/api/searchLiver?key=` + e).then(res => {
                            if (!res.data.result) return;
                            setFilters(res.data.result);
                        })
                    }}
                    onClear={() => {
                        setLiver(-1)
                    }}
                >
                    {(f) => <AutocompleteItem key={f.UID}>{f.UName}</AutocompleteItem>}
                </Autocomplete>
                <Select
                    className="max-w-xs mt-4 mb-4 ml-4"
                    items={[{
                        key: 'Time',
                        value: "Time"
                    },
                        {
                            key: 'money',
                            value: "Money"

                        },
                        {
                            key: 'message',
                            value: "Message"
                        }
                    ]}
                    label="Sort by"
                    onSelectionChange={e => {
                        setOrder(e.currentKey)
                    }}
                >
                    {(f) => <SelectItem key={f.key}>{f.value}</SelectItem>}
                </Select>
            </div>

            <div className={'grid grid-cols-1 md:grid-cols-4 2xl:grid-cols-5'}>
                {dataSource.map(item => {
                    return (
                        <LiveStatisticCard item={item} showUser/>
                    )
                })}
            </div>
            <Pagination
                isCompact
                showControls
                showShadow
                color="secondary"
                page={currentPage}
                total={total / pageSize}
                initialPage={1}
                onChange={(page) => handlePageChange(page, pageSize)}
                classNames={{
                    wrapper:'w-full mx-4',
                }}
            />
        </div>
    )
}

export default LivePage
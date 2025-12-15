import React, {useEffect, useState} from 'react';
import axios from "axios";
import {useNavigate} from "react-router";
import "./LivePage.css"
import {Autocomplete, AutocompleteItem, Button, Pagination, Select, SelectItem, useDisclosure,} from "@heroui/react";
import LiveStatisticCard from "../components/LiveStatisticCard";
import CommentForm from "../components/CommentForm";

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
const MessageIcon = (props) => (
    <svg xmlns="http://www.w3.org/2000/svg" className="icon" viewBox="0 0 1024 1024"
         style={{ width: '32px', height: '32px' }}>
        <path
            d="M464 512a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm200 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm-400 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm661.2-173.6c-22.6-53.7-55-101.9-96.3-143.3a444.35 444.35 0 0 0-143.3-96.3C630.6 75.7 572.2 64 512 64h-2c-60.6.3-119.3 12.3-174.5 35.9a445.35 445.35 0 0 0-142 96.5c-40.9 41.3-73 89.3-95.2 142.8-23 55.4-34.6 114.3-34.3 174.9A449.4 449.4 0 0 0 112 714v152a46 46 0 0 0 46 46h152.1A449.4 449.4 0 0 0 510 960h2.1c59.9 0 118-11.6 172.7-34.3a444.48 444.48 0 0 0 142.8-95.2c41.3-40.9 73.8-88.7 96.5-142 23.6-55.2 35.6-113.9 35.9-174.5.3-60.9-11.5-120-34.8-175.6zm-151.1 438C704 845.8 611 884 512 884h-1.7c-60.3-.3-120.2-15.3-173.1-43.5l-8.4-4.5H188V695.2l-4.5-8.4C155.3 633.9 140.3 574 140 513.7c-.4-99.7 37.7-193.3 107.6-263.8 69.8-70.5 163.1-109.5 262.8-109.9h1.7c50 0 98.5 9.7 144.2 28.9 44.6 18.7 84.6 45.6 119 80 34.3 34.3 61.3 74.4 80 119 19.4 46.2 29.1 95.2 28.9 145.8-.6 99.6-39.7 192.9-110.1 262.7z" />
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
    const [liver, setLiver] = useState(window.SEARCH_LIVER)
    const [commentOpen,setCommentOpen] = useState(false);

    const redirect = useNavigate()
    const refreshData = (page, size, name) => {
        var url = `${protocol}://${host}:${port}/api/live?page=` + page + "&limit=" + size
        if (liver != null && liver !== "") {
            window.SEARCH_LIVER = liver
            url = url + "&uid=" + (liver === -1 ? "0" : liver)
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

    const [currentPage, setCurrentPage] = useState(window.page ?? 1);

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
    }, [liver, order])
    useEffect(() => {
        if (chartId !== null && chartId !== 0) {
            setChart(true);
        }
    }, [chartId]); // 监听 chartId 变化后再设置 chart

    return (

        <div className={''}>
            <CommentForm isOpen={commentOpen} onChange={() => setCommentOpen(!commentOpen)} onClose={() => setCommentOpen(false)}/>
            <div className={'fixed right-[3vw] bottom-[3vw] z-40'}>
                <Button
                    isIconOnly
                    startContent={<MessageIcon/>}
                    onClick={() => {
                        setCommentOpen(true)
                    }}
                />
            </div>
            <div className='sm:flex-row flex mb-4 flex-col'>
                <Autocomplete
                    className="max-w-xs mt-4 mb-4 ml-4"
                    items={filters}
                    isClearable
                    defaultInputValue={window.LIVER_NAME ?? ''}
                    label="Liver"
                    onSelectionChange={e => {
                        setCurrentPage(1)

                        filters.forEach(filter => {
                            if (filter.UID === parseInt(e)) {
                                setLiver(filter.UID)
                                window.LIVER_NAME = filter.UName
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
                        window.SEARCH_LIVER = ''
                        setLiver(-1)
                    }}
                >
                    {(f) => <AutocompleteItem key={f.UID}>{f.UName}</AutocompleteItem>}
                </Autocomplete>
                <Select
                    isClearable
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
                    wrapper: 'w-full mx-4',
                }}
            />
        </div>
    )
}

export default LivePage
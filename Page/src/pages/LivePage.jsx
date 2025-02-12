import React, {useEffect, useState} from 'react';
import {FloatButton, Input, Table} from "antd";
import axios from "axios";
import {useNavigate} from "react-router";
import "./LivePage.css"

function LivePage(props) {


    const [dataSource, setDatasource] = useState([])

    const [total, setTotal] = useState(0)

    const [selected, isSelected] = useState(false)

    const [name, setName] = useState(null)


    const redirect = useNavigate()
    const refreshData = (page, size, name) => {
        var url = "http://localhost:8080/live?page=" + page + "&limit=" + size
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

    const [filters, setFilters] = useState([
        {text: 'Joe', value: 'Joe'},
        {text: 'Jim', value: 'Jim'},
        {text: 'Category 1', value: 'Category 1'},
        {text: 'Category 2', value: 'Category 2'},
    ]);

    const [columns, setColumn] = useState([])
    useEffect(() => {

        setColumn([
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
    }, [])
    const [currentPage, setCurrentPage] = useState(1);

    const [pageSize, setPageSize] = useState(10);

    // 处理页码改变事件
    const handlePageChange = (page, pageSize) => {
        console.log(`page=${page}  pageSize=${pageSize}`)
        refreshData(page, pageSize, name)
        setCurrentPage(page)
        setPageSize(pageSize)

    }

    setInterval(() => {
        const element = document.querySelector(".ant-table-filter-dropdown-search-input")

        if (element != null && selected === false) {
            isSelected(true)
            if (JSON.stringify(element.getEventListeners()) === `{}`) {
                const btn = document.querySelector(".ant-table-filter-dropdown-btns").childNodes[1]
                const labelGroup = document.querySelector(".ant-dropdown-menu").childNodes
                document.querySelector(".ant-table-filter-dropdown-search-input").addEventListener('input', (e) => {
                    var text = element.childNodes[1].value
                    axios.get("http://localhost:8080/liver?key=" + text).then(res => {
                        setFilters([])
                        var array = []
                        res.data.result.map((item, index) => {
                            array.push({text: item, value: item})
                        })
                        setFilters(array)
                        if (columns.length !== 0) {
                            columns[0].filters = array
                            setColumn(columns)
                        }

                    })
                })
                btn.addEventListener('click', (e) => {
                    console.log(e)
                    labelGroup.forEach((item, index) => {

                        if (item.className.indexOf('selected') !== -1) {
                            //console.log(textContent)
                            setName(item.textContent)
                            refreshData(currentPage, pageSize, item.textContent)
                        }
                    })

                })

            }
        }
    }, 50)
    return (

        <div>
            <FloatButton onClick={() => {
                axios.get("http://localhost:8080/refreshMoney").then(res => {
                    refreshData(currentPage, pageSize)
                })
            }} type="primary">Refresh Money</FloatButton>
            <Input
                placeholder="Search Filters"
                style={{marginBottom: 16}}
            ></Input>
            <Table dataSource={dataSource} columns={columns} pagination={{
                current: currentPage,             // 当前页
                total: total,
                onChange: handlePageChange,

            }}
                   onRow={(record) => {
                       return {
                           onClick: (event) => {
                               console.log(record);
                               redirect(`/lives/${record.ID}`)
                           }, // 点击行
                       };
                   }}
                   rowClassName={(record, index) => {
                       return index % 2 === 0 ? "even-row" : "odd-row"
                   }
                   }
            />
        </div>
    )
}

export default LivePage
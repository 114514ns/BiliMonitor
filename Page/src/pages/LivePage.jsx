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

    const [searchText, setSearchText] = useState("");

    const host = location.hostname;

    const redirect = useNavigate()
    const refreshData = (page, size, name) => {
        var url = `http://${host}:8080/live?page=` + page + "&limit=" + size
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
    useEffect(() => {

        //setColumn()
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
    useEffect(() => {
        console.log("columns 更新了:", columns);
    }, [columns]); // 监听 columns 变化
    useEffect(() => {
        console.log("filters 更新了:", filters);
        setColumn((prevColumns) =>
            prevColumns.map((col, index) =>
                index === 0 ? { ...col, filters: filters } : col
            )
        );
    }, [filters]);



    setInterval(() => {
        const element = document.querySelector(".ant-table-filter-dropdown-search-input")

        if (element != null && selected === false) {
            isSelected(true)
            if (JSON.stringify(element.getEventListeners()) === `{}`) {
                const btn = document.querySelector(".ant-table-filter-dropdown-btns").childNodes[1]
                const labelGroup = document.querySelector(".ant-dropdown-menu").childNodes
                document.querySelector(".ant-table-filter-dropdown-search-input").addEventListener('input', (e) => {
                    var text = element.childNodes[1].value
                    axios.get(`http://${host}:8080/liver?key=` + text).then(res => {
                        if (!res.data.result) return; // 处理 null/undefined/空数据
                        var array = []
                        const newFilters = res.data.result.map((item) => ({ text: item, value: item }));

                        setFilters(newFilters);


                    })
                })
                btn.addEventListener('click', (e) => {
                    console.log(e)
                    var found = false
                    labelGroup.forEach((item, index) => {

                        if (item.className.indexOf('selected') !== -1) {
                            //console.log(textContent)
                            setName(item.textContent)
                            console.log(item.textContent)
                            found = true

                        }

                    })
                    if (!found) {
                        refreshData(currentPage, pageSize, null)
                    }

                })

            }
        }
    }, 50)


    const onChange = (pagination, filters, sorter, extra) => {
        //refreshData(currentPage, pageSize, filters[0].value)
        console.log(pagination, filters, sorter, extra)

    };
    return (

        <div>
            <FloatButton onClick={() => {
                axios.get(`http://${host}:8080/refreshMoney`).then(res => {
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
                   onChange={onChange}
            />
        </div>
    )
}

export default LivePage
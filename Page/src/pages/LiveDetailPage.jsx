import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import axios from "axios";
import {Table} from "antd";
import  "./LivePage.css"

function LiveDetailPage(props) {
    let { id } = useParams();
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
    const refreshData = (page, size, name,order) => {
        if (page === undefined) {
            return
        }
        var url = "http://localhost:8080/live/" + id + "/?" +  "page=" + page + "&limit=" + size + "&order=" + order
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
    const onChange = (pagination, filters, sorter, extra) => {
        console.log(pagination)

        refreshData(pagination.current,pagination.pageSize,null,sorter.order)
    };



    return (
        <div>
            <Table dataSource={dataSource} columns={columns} pagination={{
                current: currentPage,             // 当前页
                total: total,
                onChange: handlePageChange,

            }}
                   onRow={(record) => {
                       return {
                           onClick: (event) => {}, // 点击行
                       };
                   }}
                   onChange={onChange}
                   rowClassName={(record, index) => {
                       return index % 2 === 0 ? "even-row" : "odd-row"
                   }
                   }
            />
        </div>
    );
}

export default LiveDetailPage;
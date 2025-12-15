import React, { useEffect } from 'react';
import {
    Avatar,
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
import { CheckIcon } from "../pages/ChatPage";
import HoverMedals from "./HoverMedals";
import { NavLink, useNavigate } from "react-router-dom";
import axios from "axios";

function ActionTable(props) {

    const [currentPage, setCurrentPage] = React.useState(window.USER_PAGE ? window.USER_PAGE : 1);


    const redirect = useNavigate();

    const pageSize = parseInt(localStorage.getItem("defaultPageSize"))

    useEffect(() => {
        props.handlePageChange(currentPage)
    }, [])

    const columns = [
        {
            title: 'Name',
            dataIndex: 'FromName',
            key: 'UserName',
            render: (text, record) => (
                <span style={{ cursor: 'pointer' }} onClick={() => {
                    window.open("https://space.bilibili.com/" + record.FromId)
                }}>
                    {text}{console.log(record)}
                </span>
            )
        },
        {
            title: 'Liver',
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
            sorter: true,

        },
        {
            title: 'Message',
            dataIndex: 'Extra',
            key: 'Message'
        }
    ]

    const tableRef = React.createRef();

    return (
        <div className={''}>
            <Table bottomContent={
                <div className="flex w-full justify-center">
                    <Pagination
                        isCompact
                        showControls
                        showShadow
                        color="secondary"
                        page={currentPage}
                        initialPage={window.USER_PAGE ? window.USER_PAGE : 1}
                        total={Math.ceil(props.total / pageSize)}
                        onChange={(page) => {
                            props.handlePageChange(page, pageSize)

                        }}
                    />
                </div>
            } isStriped ref={tableRef}>

                <TableHeader>
                    {columns.map((col, index) => (
                        <TableColumn key={index}>{col.title}</TableColumn>

                    ))}
                </TableHeader>
                <TableBody>

                    {props.dataSource?.map((item, index) => (
                        <TableRow key={index} onClick={() => {

                        }}>
                            <TableCell>
                                <div className={'hover:text-gray-500 flex items-center'} onClick={() => {
                                    toSpace(item.FromId)
                                }}>
                                    {item.FromName}
                                    {item.MedalLevel != 0 ? <Chip
                                        className={'basis-64'}
                                        startContent={<CheckIcon size={18} />}
                                        variant="faded"
                                        onClick={() => {
                                            toSpace(item.MedalLiver);
                                        }}
                                        style={{ background: getColor(item.MedalLevel), color: 'white', marginLeft: '8px' }}
                                    >
                                        {item.MedalName}
                                        <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                            {item.MedalLevel}
                                        </span>
                                    </Chip> : <></>}
                                </div>
                            </TableCell>
                            <TableCell>
                                <NavLink className={'flex hover:text-gray-500'} to={"/liver/" + item.UserID}>
                                    {!isMobile() && <Avatar
                                        className={'w-[40px] h-[40px]'}
                                        src={`${AVATAR_API}${item.UserID}`}
                                    />}
                                    <span className={'ml-2 mt-2'}>
                                        {item.UserName}
                                    </span>
                                </NavLink>
                            </TableCell>
                            <TableCell>{new Date(item.CreatedAt).toLocaleString()}</TableCell>
                            <TableCell>{item.GiftPrice.Float64}</TableCell>
                            <TableCell onClick={() => {
                                axios.get(`${protocol}://${host}:${port}/api/queryPage?id=${item.ID}&live=${item.Live}`).then((response) => {
                                    redirect(`/lives/${item.Live}?page=${response.data.page}&highLight=${item.ID}`);
                                });
                            }}>
                                <div className={'transition-transform hover:text-gray-500'}>
                                    <Tooltip content={'点击跳转'} placement='left'>
                                        <span>
                                            {item.GiftName || item.Extra}
                                        </span>
                                    </Tooltip>

                                    {item.ActionName === "gift" && item.GiftAmount.Int16 !== 1 && <span className={'font-bold'}>*{item.GiftAmount.Int16}</span>}
                                </div>

                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

export default ActionTable;
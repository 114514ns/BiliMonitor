import React, {useEffect} from 'react';
import {
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
import {CheckIcon} from "../pages/ChatPage";
import HoverMedals from "./HoverMedals";

function ActionTable(props) {

    const [currentPage, setCurrentPage] = React.useState(1);



    const pageSize = 10

    useEffect(() => {
        props.handlePageChange(1.10)
    },[])

    const columns = [
        {
            title: 'Name',
            dataIndex: 'FromName',
            key: 'UserName',
            render: (text, record) => (
                <span style={{cursor: 'pointer'}} onClick={() => {
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

    return (
        <div>
            <Table bottomContent={
                <div className="flex w-full justify-center">
                    <Pagination
                        isCompact
                        showControls
                        showShadow
                        color="secondary"
                        page={currentPage}
                        total={Math.ceil(props.total / pageSize)}
                        onChange={(page) => props.handlePageChange(page, pageSize)}
                    />
                </div>
            } isStriped>

                <TableHeader>
                    {columns.map((col, index) => (
                        <TableColumn key={index}>{col.title}</TableColumn>

                    ))}
                </TableHeader>
                <TableBody>

                    {props.dataSource.map((item, index) => (
                        <TableRow key={index} onClick={() => {

                        }}>
                            <TableCell>
                                <Tooltip content={
                                    <HoverMedals mid={item.FromId}/>
                                } delay={400}>
                                    <div className={'flex'} onClick={() => {
                                        toSpace(item.FromId)
                                    }}>
                                        {item.FromName}
                                        {item.MedalLevel != 0 ?                                     <Chip
                                            className={'basis-64'}
                                            startContent={<CheckIcon size={18}/>}
                                            variant="faded"
                                            onClick={() => {
                                                toSpace(item.MedalLiver);
                                            }}
                                            style={{background: getColor(item.MedalLevel), color: 'white', marginLeft: '8px'}}
                                        >
                                            {item.MedalName}
                                            <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {item.MedalLevel}
                                                        </span>
                                        </Chip>:<></>}
                                    </div>
                                </Tooltip>
                            </TableCell>
                            <TableCell>{item.ToName}</TableCell>
                            <TableCell>{item.CreatedAt}</TableCell>
                            <TableCell>{item.GiftPrice.Float64}</TableCell>
                            <TableCell>{item.GiftName || item.Extra }{item.ActionName==="gift" && item.GiftAmount.Int16 !== 1 && <span className={'font-bold'}>*{item.GiftAmount.Int16}</span>}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

export default ActionTable;
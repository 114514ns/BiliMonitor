import React, {useEffect, useState} from 'react';
import {
    Input,
    Modal,
    ModalBody,
    ModalContent,
    ModalHeader,
    TableColumn,
    TableHeader,
    Table,
    TableRow, TableCell, Avatar, TableBody, Chip
} from "@heroui/react";
import axios from "axios";

function MysteryBoxStatistic(props) {

    const [data,setData] = useState([])

    useEffect(() => {
        if (props.isOpen) {
            axios(`/api/box/${props.type}?uid=${props.uid}`).then((res) => {
                console.log(res.data)
                setData(res.data.data)
            })
        }
    },[props.isOpen])

    return (
        <Modal isOpen={props.isOpen} onOpenChange={props.onClose} className={'overflow-scroll scrollbar-hide'} scrollBehavior={'inside'} size={'2xl'}>
            <ModalContent>
                <ModalHeader>盲盒统计</ModalHeader>
                <ModalBody>
                    <div className={'flex flex-col'}>
                        <Table>
                            <TableHeader>
                                <TableColumn>{props.type === 'user'?'Liver':'Watcher'}</TableColumn>
                                <TableColumn>Rate</TableColumn>
                                <TableColumn>BoxCount</TableColumn>
                                <TableColumn>Spend</TableColumn>
                            </TableHeader>
                            <TableBody>
                                {data.map((e) => {
                                    return (
                                        <TableRow>
                                            <TableCell className={'flex flex-row items-center '} onClick={() => {
                                                window.open(props.type === 'user'?`/liver/${e.LiverUID}`:`/user/${e.UID}`)
                                            }}>
                                                <Avatar src={`${AVATAR_API}${props.type === 'user'?e.LiverUID:e.UID}`}/>
                                                <p className={'ml-2'}>{props.type === 'user'?e.LiverName:e.UName}</p>
                                            </TableCell>
                                            <TableCell>
                                                <Chip variant={'flat'} color={e.ReturnRate > 1?'success':'danger'}>{Math.trunc(e.ReturnRate*100)}%</Chip>
                                            </TableCell>
                                            <TableCell className={'font-bold'}>
                                                {e.Count}
                                            </TableCell>
                                            <TableCell className={'font-bold'}>
                                                {e.Spend}
                                            </TableCell>
                                        </TableRow>
                                    )
                                })}
                            </TableBody>
                        </Table>
                    </div>
                </ModalBody>
            </ModalContent>
        </Modal>
    );
}

export default MysteryBoxStatistic;
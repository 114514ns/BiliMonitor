import React, {useEffect} from 'react';
import axios from "axios";
import {
    Chip,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow,
    Table,
    Button,
    Input,
    Modal, ModalContent, ModalBody, ModalHeader, ModalFooter, useDisclosure, Alert
} from "@heroui/react";
import {NavLink} from "react-router-dom";

const MailIcon = () => {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" className="icon" viewBox="0 0 1024 1024"     style={{ width: '32px', height: '32px' }}>>
            <path d="M928 160H96c-17.7 0-32 14.3-32 32v640c0 17.7 14.3 32 32 32h832c17.7 0 32-14.3 32-32V192c0-17.7-14.3-32-32-32zm-40 110.8V792H136V270.8l-27.6-21.5 39.3-50.5 42.8 33.3h643.1l42.8-33.3 39.3 50.5-27.7 21.5zM833.6 232L512 482 190.4 232l-42.8-33.3-39.3 50.5 27.6 21.5 341.6 265.6a55.99 55.99 0 0 0 68.7 0L888 270.8l27.6-21.5-39.3-50.5-42.7 33.2z"/>
        </svg>
    )
}

function TracePage(props) {
    const [data, setData] = React.useState([]);

    const refresh = () => {
        axios.get("/api/trace_srv/list").then(res => {
            setData((res.data.list ?? []).sort((a,b) => {
                return a.Allow - b.Allow
            }))
        })
    }
    useEffect(() => {
        refresh()
    }, [])
    const uname = React.createRef()
    const fans = React.createRef()
    const [obj,setObj] = React.useState({})
    const [uid, setUid] = React.useState('')
    const {isOpen, onOpen, onOpenChange} = useDisclosure();
    return (
        <div>
            <Modal isOpen={isOpen} onOpenChange={onOpenChange}>
                <ModalContent>
                    {(onClose) => (
                        <>
                            <ModalHeader className="flex flex-col gap-1">Confirm</ModalHeader>
                            <ModalBody>
                                    <div className="flex flex-col items-center">
                                        <img
                                            src={`${AVATAR_API}${uid}`}
                                            className="h-[40px] w-[40px] rounded-full"
                                        />
                                        <p>{uname.current}</p>
                                    </div>
                                <Alert color={obj.Fans > 1000?'warning':'danger'} >
                                    {obj.Fans>1000?
                                        obj.Fans>4000?<p>4000粉丝以上的虚拟直播已经可以稳定记录，请先<a href={`/liver/${uid}`} className={'text-blue-500'}>确认</a>是否需要添加</p>:<p></p>:
                                        <div className={'flex flex-row items-center'}>
                                            不接受臭底边喵<img src={
                                            'https://i0.hdslb.com/bfs/new_dyn/f5d72fdaa70520847381b8bb7ef531941995486878.png'} className={'ml-3 w-[60px] h-[60px]'}></img>
                                        </div>
                                    }
                                </Alert>
                            </ModalBody>
                            <ModalFooter>
                                <Button color="danger" variant="light" onPress={onClose}>
                                    Close
                                </Button>
                                <Button color="primary" onPress={() => {
                                    axios.post("/api/trace_srv/submit",new URLSearchParams({
                                        uid: uid,
                                    })).then(() => {
                                        refresh()
                                        onClose()
                                    })
                                }} isDisabled ={obj.Fans < 1000}>
                                    Confirm
                                </Button>
                            </ModalFooter>
                        </>
                    )}
                </ModalContent>
            </Modal>
            <div className={'flex flex-row mb-4 items-center'}>
                <Input label="Input UID" className={'max-w-xs'} value={uid} onValueChange={(e) => {setUid(e.replace('UID:',''))}}/>
                <Button className={'ml-4'} onClick={() => {
                    axios.get(`/api/trace_srv/info?mid=${uid}`).then(res => {
                        setObj(res.data)
                        uname.current = res.data.UName
                        fans.current = res.data.Fans
                        onOpen()
                    })
                }}>Submit</Button>
            </div>
            <Table aria-label="Example static collection table">
                <TableHeader>
                    <TableColumn>Liver</TableColumn>
                    <TableColumn>State</TableColumn>
                </TableHeader>
                <TableBody>
                    {data.map((item) => {
                        return (
                            <TableRow key={item.UID}>
                                <TableCell>
                                    <div className="flex flex-row items-center hover:text-gray-500">
                                        <img
                                            src={`${AVATAR_API}${item.UID}`}
                                            className="h-[40px] w-[40px] rounded-full"
                                        />
                                        <NavLink className="ml-2" to={`/liver/${item.UID}`}>{item.UName}</NavLink>
                                    </div>
                                </TableCell>
                                <TableCell onClick={() => {
                                    if (!import.meta.env.PROD) {
                                        if (!item.Allow) {
                                            axios.post("/api/trace_srv/allow",new URLSearchParams({
                                                "room":item.Room,
                                            })).then(() => {
                                                refresh()
                                            })
                                        }

                                    }
                                }}>
                                    {item.Allow?<Chip color={'success'} variant={'flat'}>Active</Chip>:<Chip color={'danger'} variant={'flat'}>Reviewing</Chip>}
                                </TableCell>
                            </TableRow>
                        )
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

export default TracePage;
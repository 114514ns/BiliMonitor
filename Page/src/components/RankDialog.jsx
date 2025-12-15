import React, {memo, useEffect, useMemo} from 'react';
import {
    Autocomplete,
    AutocompleteItem,
    Avatar,
    Chip,
    Listbox,
    ListboxItem,
    Modal,
    ModalBody,
    ModalContent,
    ModalHeader,
    Pagination, SelectItem,
    Switch, Tooltip,Dropdown,DropdownMenu,DropdownItem
} from "@heroui/react";
import axios from "axios";
import {CheckIcon} from "../pages/ChatPage";
import {useNavigate} from "react-router";
import UserChip from "./UserChip";

window.getColor = (level) => {
    if (level <= 10) {
        return "#727BB5"
    }
    if (level <= 20) {
        return "#CF86B2"
    }
    if (level <= 30) {
        return "#5EC0F7"
    }
    if (level <= 40) {
        return "#6992FF"
    }
    if (level <= 50) {
        return "#AA78F1"
    }
    if (level <= 60) {
        return "#ED5674"
    }
    if (level <= 70) {
        return "#F58737"
    }
    if (level <= 80) {
        return "#F58837"
    }
    if (level <= 90) {
        return "#F58837"
    }
    if (level <= 40) {
        return "#FF9D55"
    }


}
const ref = React.createRef();
function RankDialog(props) {

    const [data, setData] = React.useState([]);

    window.redirect = useNavigate()

    const fetchData = (key) => {
        if (!key) {
            key = "1"
        }
        axios.get("/api/searchAreaLiver?key=" + key).then((response) => {
            setData(response.data.result);
        })
    }
    const [fans, setFans] = React.useState([]);

    const [page, setPage] = React.useState(1);

    const [liver, setLiver] = React.useState("");

    const [totalPage, setTotalPage] = React.useState(1);





    const switchRef = React.createRef();

    const [activeOnly, setActiveOnly] = React.useState(false);

    const SIZE = 100

    const fetchFans = () => {
        axios.get("/api/fansRank?liver=" + liver?.split("-")[0] + `&size=${SIZE}&page=` + page).then((response) => {
            setFans(response.data.list);
            if (response.data.total % SIZE === 0) {
                setTotalPage(response.data.total / SIZE);
            } else {
                setTotalPage(Math.floor(response.data.total / SIZE) + 1);
            }

            console.log(switchRef.current);
            if (ref.current) {
                const element = ref.current.children[0].children[0]
                element.scrollTop = 0;
                console.log("top ", element.scrollTop, "height ", element.height);
            }
        })
    }

    useEffect(() => {
        fetchFans()
    }, [page, liver])

    useEffect(() => {
        fetchData();
    }, [])
    return (
        <div>
            <Modal isOpen={props.open} onClose={() => {
                setData([])

                props.onClose();
            }} size="lg" >
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">Rank</ModalHeader>
                    <ModalBody>

                        <div className="flex flex-col">
                            <div className={'flex flex-row gap-1'}>
                                <Autocomplete
                                    className="max-w-xs"
                                    defaultItems={data}
                                    label="Search"
                                    labelPlacement="inside"
                                    placeholder="Select a user"
                                    onInputChange={(value) => {
                                        fetchData(value);
                                    }}
                                    onSelectionChange={(value) => {
                                        setLiver(value)
                                        setPage(1)

                                    }}
                                >
                                    {(user) => (
                                        <AutocompleteItem key={`${user.UID}-${user.LiverID}`} textValue={user.UName}>
                                            <div className="flex gap-2 items-center">
                                                <Avatar
                                                    className="flex-shrink-0"
                                                    size="sm"
                                                    src={`${AVATAR_API}${user.UID}`}
                                                ></Avatar>

                                                <div className="flex flex-col">
                                                    <span className="text-small">{user.UName}</span>
                                                    <span className="text-tiny text-default-400">{user.UName}</span>
                                                </div>
                                            </div>
                                        </AutocompleteItem>
                                    )}
                                </Autocomplete>
                                <Switch
                                    className='ml-2 '
                                    ref={switchRef}
                                >Active Only</Switch>
                            </div>

                            <FansList fans={fans} onClose={props.onClose} inspect={props.inspect} height={520}/>

                        </div>

                        <Pagination initialPage={1} total={totalPage} className='content-center' onChange={e => {
                            setPage(e)
                        }}/>
                    </ModalBody>
                </ModalContent>
            </Modal>
        </div>
    );
}

export const FansList = memo(function FansList({fans,onClose,height,onItemClick,inspect}) {

    const [open, setOpen] = React.useState(false);
    const [id,setId] = React.useState(0);
    const getStyle = (e) => {
        if (e === "add") {
            return 'bg-green-200 rounded-lg px-2 py-2'
        }
        if (e === "remove") {
            return 'bg-red-200 rounded-lg px-2 py-2'
        }
        return ""
    }
    return <Listbox
        style={{
            //maxHeight: "800px",
            //"overflow-y": "scroll",
        }}
        isVirtualized
        ref={ref}
        virtualization={
            {
                maxListboxHeight: height??600,
                itemHeight: 80
            }

        }
        onClick={(e) => {
            setOpen(false)
        }}
        onMouseLeave={() => setOpen(false)}
    >
        {fans.map((f) => (
            <ListboxItem
                key={f.UID + '-' + f.LiverID}
                classNames={'py-2'}
            >
                <Tooltip content={
                    <div>
                        {<Dropdown>
                            <DropdownMenu aria-label="Static Actions">
                                <DropdownItem key="new" onClick={() => {
                                    window.open("https://space.bilibili.com/" + f.UID)
                                }}>Bilibili</DropdownItem>
                                <DropdownItem key="copy" onClick={() => {
                                    window.open("/user/" + f.UID)
                                }}>KUN</DropdownItem>
                            </DropdownMenu>
                        </Dropdown>}
                    </div>
                } isOpen={(open && (id === f.UID + '-' + f.LiverID))}>
                    <div className={getStyle(f.Label)} onContextMenu={(e) => {
                        e.preventDefault();
                        setId(f.UID + '-' + f.LiverID)
                        setOpen(!open);
                    }}>
                        <p className={`text-medium ${(inspectGuard(f) && inspect)?'line-through':''}`}>{f.UName}</p>
                        {(
                            <div className={'flex flex-row align-middle '} onClick={() => {
                                onItemClick(f)
                            }}>
                                <UserChip props={convert(f)}/>

                            </div>
                        )}
                    </div>
                </Tooltip>
            </ListboxItem>
        ))}
    </Listbox>
});

const convert = (item) => {
    item.MedalLevel = item.Level
    if (!item.FromId) {
        item.FromId = item.UID;
    }
    if (item.Type) {
        item.GuardLevel = item.Type
    }
    return item;
}


export default RankDialog;


export const MoneyRankDialog = (props) => {


    const array = JSON.parse(localStorage.getItem("money"))

    const PAGE_SIZE = 100
    
    const [page,setPage] = React.useState(1)

    const [data,setData] = React.useState([])


    const totalPage = array.length/PAGE_SIZE


    useEffect(() => {

        const start = (page-1)*PAGE_SIZE

        var tmp = []

        for(var i = start;i<start+PAGE_SIZE;i++) {
            tmp.push(array[i])
        }
        setData(tmp)
    },[page])
    

    return (
        <div>
            <Modal isOpen={props.open} onClose={() => {
                setData([])

                props.onClose();
            }} size="md">
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">Money Rank</ModalHeader>
                    <ModalBody>

                        <div className="flex flex-col">
                            <FansList fans={data} onClose={props.onClose} height={isMobile()?600:900}/>
                        </div>

                        <Pagination initialPage={1} total={totalPage} className='content-center' onChange={e => {
                            setPage(e)
                        }}/>
                    </ModalBody>
                </ModalContent>
            </Modal>
        </div>
    );
}
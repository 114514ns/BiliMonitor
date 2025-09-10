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
    Pagination,
    Switch
} from "@heroui/react";
import axios from "axios";
import {CheckIcon} from "../pages/ChatPage";
import {useNavigate} from "react-router";
import UserChip from "./UserChip";

window.getColor = (level) => {
    if (level <= 4) {
        return "#5762A799"
    }
    if (level <= 8) {
        return "#5866C799"
    }
    if (level <= 12) {
        return "#9BA9EC"
    }
    if (level <= 16) {
        return "#DA9AD8"
    }
    if (level <= 20) {
        return "#C79D24"
    }
    if (level <= 24) {
        return "#67C0E7"
    }
    if (level <= 28) {
        return "#6C91F2"
    }
    if (level <= 32) {
        return "#A97EE8"
    }
    if (level <= 36) {
        return "#C96B7E"
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
            }} size="md">
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
                                                    src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${user.UID}`}
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

                            <FansList fans={fans} onClose={props.onClose} />

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

export const FansList = memo(function Greeting({fans,onClose,height,onItemClick}) {

    const getStyle = (e) => {
        if (e === "add") {
            return 'bg-green-200 rounded-lg px-2'
        }
        if (e === "remove") {
            return 'bg-red-200 rounded-lg px-2'
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
                maxListboxHeight: height??400,
                itemHeight: 80
            }

        }
    >
        {fans.map((f) => (
            <ListboxItem
                key={f.UID + '-' + f.LiverID}
            >
                <div className={getStyle(f.Label)}>
                    <p className={'font-medium'}>{f.UName}</p>
                    {(
                        <div className={'flex flex-row align-middle mt-2'} onClick={() => {
                            onItemClick(f)
                        }}>
                            <UserChip props={convert(f)}/>
        
                        </div>
                    )}
                </div>
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

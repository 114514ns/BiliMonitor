import React, {useEffect} from 'react';
import {
    Autocomplete,
    AutocompleteItem,
    Avatar,
    Button,
    Chip,
    Input,
    ScrollShadow,
    Select,
    SelectItem
} from "@heroui/react";
import axios from "axios";
import {CheckIcon} from "./ChatPage";
import {NavLink} from "react-router-dom";






function uniqueByKey(arr, key) {
    const seen = new Set();
    return arr.filter(obj => {
        if (seen.has(obj[key])) {
            return false;
        }
        seen.add(obj[key]);
        return true;
    });
}


function SearchPage(props) {

    const [type, setType] = React.useState('name')

    const [items, setItems] = React.useState([])

    const [text,setText] = React.useState("")

    const [avatar,setAvatar] = React.useState("")

    const [rooms,setRooms] = React.useState([])


    useEffect(() => {
        setAvatar("https://i1.hdslb.com/bfs/face/5ddddba98f0265265662a8f7d5383e528a98412b.jpg")
        axios.get("/api/status").then((response) => {
            setRooms(uniqueByKey(response.data.Rooms.sort((a,b) => a.Fans < b.Fans ? 1 : -1),"UID"));
        })
    }, []);

    return (

        <div className={'w-full flex flex-col items-center  h-full'}>
            <Avatar src={avatar}
                    className={'w-[200px] h-[200px] mt-[5vh]'}>

            </Avatar>
            <ScrollShadow className={'w-[80vw] sm:w-[50vw] max-h-[20vh] overflow-scroll scrollbar-hide mt-6'}>

                    {rooms.map((room, index) => (
                        <NavLink to={'/liver/' + room.UID}>
                            <Chip
                                avatar={<Avatar name={room.UName} src={room.Face} />}
                                variant="flat"
                                className={'ml-2 mt-1'}
                            >
                                <p className={'font-bold'}>{room.UName}</p>
                            </Chip>
                        </NavLink>))}
            </ScrollShadow>
            <div className={'flex w-full mt-[6vh] sm:flex-row flex-col'}>
                <Select className="sm:max-w-xs sm:mr-4 " onSelectionChange={(e) => {
                    setType(e.currentKey);
                }} label={'Type'} selectedKeys={['name']}>
                    <SelectItem key={'room'}>Room</SelectItem>
                    <SelectItem key={'uid'}>UID</SelectItem>
                    <SelectItem key={'name'}>UName</SelectItem>
                    <SelectItem key={'watcher-name'}>Watcher</SelectItem>
                </Select>
                {(type === 'name' || type === 'watcher-name') &&
                    <Autocomplete className="sm:ml-4 sm:mt-0 mt-4" label="Select..." onInputChange={(e) => {
                        if (e !== '') {
                            axios.get(`/api/search?type=${type}&key=${e}&api=1`).then((response) => {
                                if (response.data.data != null) {
                                    if (response.data.data.length >= 1) {
                                        if (response.data.data[0].UName !== '') {
                                            setItems(response.data.data ?? []);
                                        }
                                    }
                                }

                            })
                        }
                    }}>
                        {items.map((e) => (
                            <AutocompleteItem key={String(e.UID)} textValue={e.UName} onClick={() => {
                                if (type === "watcher-name") {
                                    window.open("/user/" + e.UID)
                                }
                                if (type === "name") {
                                    window.open("/liver/" + e.UID)
                                }
                            }} onMouseEnter={() => {
                                setAvatar(`${AVATAR_API}${e.UID}`)
                            }}>
                                <div className={'flex flex-row'}>
                                    <div className={'flex flex-col'}>
                                        <Avatar src={`${AVATAR_API}${e.UID}`} className={'mr-2'}/>

                                    </div>
                                    <div>
                                        <p className={'font-bold'}>{e.UName}</p>
                                        {type === "name" && <p className={' text-small'}>
                                            {e.ExtraInt.toLocaleString()} Fans
                                        </p>}
                                        {type === "watcher-name" && <p className={' text-small'}>
                                            <Chip
                                                startContent={<CheckIcon size={18}/>}
                                                variant="faded"
                                                onClick={() => {
                                                    toSpace(e.LiverID);
                                                }}
                                                style={{
                                                    background: getColor(e.MedalLevel),
                                                    color: 'white',
                                                    marginLeft: '8px',
                                                    marginTop: '4px'
                                                }}
                                            >
                                                {e.MedalName}
                                                <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {e.MedalLevel}
                                                        </span>
                                            </Chip>
                                        </p>}
                                    </div>

                                </div>
                            </AutocompleteItem>
                        ))}
                    </Autocomplete>}
                {(type === 'room' || type.includes('id')) &&
                    <div className={'flex w-full items-center'}>
                        <Input label={'Input...'} onChange={(e) => {
                            setText(e.target.value)
                        }}>
                        </Input>
                        <Button color={'primary'} onPress={() => {
                            if (type === "room") {
                                axios.get(`/api/search?type=${type}&key=${text}&api=1`).then((response) => {
                                    if (response.data.data != null) {
                                        if (response.data.data.length >= 1) {
                                            if (response.data.data[0].UName !== '') {
                                                window.open("/liver/" + response.data.data[0].UID )
                                            }
                                        }
                                    }

                                })
                            }
                            if (type === "uid") {
                                window.open("/liver/" + text)
                            }
                        }} className={'ml-2'}>
                            Go
                        </Button>
                    </div>

                }
            </div>

        </div>
    );
}

export default SearchPage;
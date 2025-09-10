import React, {useEffect} from 'react';
import {Autocomplete, AutocompleteItem, Avatar, Button, Chip, Input, Select} from "@heroui/react";
import axios from "axios";
import {CheckIcon} from "./ChatPage";


function SearchPage(props) {

    const [type, setType] = React.useState('name');

    const [items, setItems] = React.useState([]);

    const [text,setText] = React.useState("");

    const [avatar,setAvatar] = React.useState("");

    useEffect(() => {
        setAvatar("https://i1.hdslb.com/bfs/face/5ddddba98f0265265662a8f7d5383e528a98412b.jpg")
    }, []);

    return (
        <div className={'w-full flex flex-col items-center  h-full'}>
            <Avatar src={avatar}
                    className={'w-[200px] h-[200px] mt-[5vh]'}>

            </Avatar>
            <div className={'flex w-full mt-[20vh] sm:flex-row flex-col'}>
                <Select className="max-w-xs sm:mr-4 " onSelectionChange={(e) => {
                    setType(e.currentKey);
                }} label={'Type'} defaultSelectedKeys={'name'}>
                    <AutocompleteItem key={'room'}>Room</AutocompleteItem>
                    <AutocompleteItem key={'uid'}>UID</AutocompleteItem>
                    <AutocompleteItem key={'name'}>UName</AutocompleteItem>
                    <AutocompleteItem key={'watcher-name'}>Watcher</AutocompleteItem>
                </Select>
                {(type === 'name' || type === 'watcher-name') &&
                    <Autocomplete className="sm:ml-4" label="Select..." onInputChange={(e) => {
                        if (e !== '') {
                            axios.get(`/api/search?type=${type}&key=${e}`).then((response) => {
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
                                setAvatar(`${import.meta.env.PROD ? '/' : '/api'}face?mid=${e.UID}`)
                            }}>
                                <div className={'flex flex-row'}>
                                    <div>
                                        <Avatar src={`${import.meta.env.PROD ? '/' : '/api'}face?mid=${e.UID}`} className={'mr-2'}/>
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
                                axios.get(`/api/search?type=${type}&key=${text}`).then((response) => {
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
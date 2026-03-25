import React, {useEffect} from 'react';
import {
    Autocomplete,
    AutocompleteItem,
    Avatar,
    Chip, Dropdown,
    DropdownItem,
    DropdownMenu, ScrollShadow,
    Select,
    SelectItem,
    Tooltip
} from "@heroui/react";
import axios from "axios";
import {CheckIcon} from "./ChatPage";
import {NavLink, useNavigate} from "react-router-dom";
import { useLongPress } from 'react-use'

function isNumber(str) {
    return !isNaN(Number(str)) && str.trim() !== '';
}


function EyeIcon(props) {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="128" height="128">
            <defs>
                <path id="eyePath" d="M 10 50 Q 50 10 90 50 Q 50 90 10 50">
                    <animate attributeName="d"
                             values="M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 48 90 50 Q 50 52 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 48 90 50 Q 50 52 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 48 90 50 Q 50 52 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 48 90 50 Q 50 52 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50;
                         M 10 50 Q 50 10 90 50 Q 50 90 10 50"
                             keyTimes="0; 0.15; 0.17; 0.19; 0.45; 0.47; 0.49; 0.51; 0.53; 0.85; 0.87; 0.89; 1"
                             dur="10s" repeatCount="indefinite" />
                </path>
                <clipPath id="eyeClip">
                    <use href="#eyePath" />
                </clipPath>
            </defs>

            <use href="#eyePath" fill="#ffffff" stroke="#333333" strokeWidth="4" strokeLinejoin="round"/>

            <g clipPath="url(#eyeClip)">
                <g>
                    <circle cx="50" cy="50" r="18" fill="#990000" />
                    <circle cx="50" cy="50" r="13" fill="#CC0000" />
                    <circle cx="50" cy="50" r="7" fill="#111111" />
                    <circle cx="43" cy="43" r="3.5" fill="#ffffff" />
                </g>
            </g>

            <use href="#eyePath" fill="none" stroke="#333333" strokeWidth="4" strokeLinejoin="round"/>
        </svg>
    )
}
import { useRef, useCallback } from 'react'
const useLongPressWithClick = ({
                                   onLongPress,
                                   onClick,
                                   delay = 500,
                               }) => {
    const onLongPressRef = useRef(onLongPress)
    const onClickRef = useRef(onClick)
    useEffect(() => { onLongPressRef.current = onLongPress }, [onLongPress])
    useEffect(() => { onClickRef.current = onClick }, [onClick])
    const stateMap = useRef(new Map())
    const bind = useCallback((el) => {
        if (!el) return
        if (stateMap.current.has(el)) return
        const state = {
            timer: null,
            isLongPress: false,
            isMoved: false,
        }
        stateMap.current.set(el, state)
        const onTouchStart = (e) => {
            state.isMoved = false
            state.isLongPress = false
            state.timer = setTimeout(() => {
                if (!state.isMoved) {
                    state.isLongPress = true
                    const touch = e.touches[0]
                    onLongPressRef.current(el, touch.clientX, touch.clientY)
                }
            }, delay)
        }
        const onTouchMove = () => {
            state.isMoved = true
            clearTimeout(state.timer)
        }
        const onTouchEnd = () => {
            clearTimeout(state.timer)
        }
        const handleClick = () => {
            if (state.isLongPress) {
                state.isLongPress = false
                return
            }
            onClickRef.current?.(el)
        }
        el.addEventListener('touchstart', onTouchStart, { passive: true })
        el.addEventListener('touchmove', onTouchMove, { passive: true })
        el.addEventListener('touchend', onTouchEnd, { passive: true })
        el.addEventListener('click', handleClick)
        return () => {
            el.removeEventListener('touchstart', onTouchStart)
            el.removeEventListener('touchmove', onTouchMove)
            el.removeEventListener('touchend', onTouchEnd)
            el.removeEventListener('click', handleClick)
            stateMap.current.delete(el)
        }
    }, [delay])
    return { bind }
}
function IndexPage(props) {

    const saveConfig = () => {
        localStorage.setItem('search_history',JSON.stringify(history.filter(item => item.Pinned).concat(history.filter(item => !item.Pinned).slice(0,10))))
    }
    const loadConfig = () => {
        var obj = localStorage.getItem('search_history')
        if (obj) {
            obj = JSON.parse(localStorage.getItem('search_history'))
        } else {
            obj = [
                {
                    'UName':'麻尤米mayumi',
                    "UID":1265605287,
                    'Pinned':false,
                    'Type':'Liver'
                }
            ]
        }
        setHistory(obj.filter((item) => item.Pinned).concat(obj.filter((item) => !item.Pinned)))
    }

    const [type,setType] = React.useState('Liver');
    const [text,setText] = React.useState('');

    const [data,setData] = React.useState([])

    const [history,setHistory] = React.useState([])

    useEffect(() => {
        loadConfig()
    }, []);

    useEffect(() => {
        saveConfig()
    },[history])



    const { bind } = useLongPressWithClick({
        onLongPress: (e, x, y) => {
            var depth = 10
            var e = e.target
            while (depth > 0) {
                if (e.attributes.getNamedItem('data-uid')) {
                    const uid = parseInt(e.attributes.getNamedItem('data-uid').value)
                    setHistory(prevState => {
                        return [{
                            UID:uid,
                            UName:e.innerText,
                            Pinned:!(e.attributes.getNamedItem('data-pinned').value === 'true'),
                            Type:type,
                        },...prevState.filter(item => item.UID !== uid)]
                    })
                    break
                } else {
                    e = e.parentElement
                    depth--
                }
            }
        },
        onClick: (el) => {
            console.log('点击了', el)
        },
    })

    useEffect(() => {
        console.log(type)
        if (type === 'Watcher') { //观众
            if (isNumber(text) || isNumber(text.replace("UID:",""))) { //UID逻辑
                axios.get("/api/search/uid?type=Watcher&uid=" + text.replace("UID:","")).then((res) => {
                    setData(res.data.data.UID?[
                        {
                            MedalType:res.data.data.Type,
                            "Type":"Watcher",
                            UName:res.data.data.UName,
                            MedalLevel:res.data.data.Level,
                            MedalName:res.data.data.MedalName,
                            UID:text.replace("UID:","")
                        }
                    ]:[])
                })
            } else { //用户名模糊搜索
                axios.get("/api/search?api=1&type=watcher-name&key=" + text).then((res) => {
                    setData(res.data.data.map((item) => {
                        item.MedalType = item.Type
                        item.Type = 'Watcher'

                        return item
                    }))
                })
            }
        } else {
            //主播
            if (isNumber(text) || isNumber(text.replace("UID:",""))) { //UID逻辑

                axios.get("/api/search/uid?type=Liver&uid=" + text.replace("UID:","")).then((res) => {
                    setData(res.data.name?[{
                        Type:'Liver',
                        UName:res.data.name,
                        Fans:res.data.fans,
                        UID:text.replace("UID:","")
                    }]:[])
                })
            } else {
                axios.get('/api/search?api=1&type=name&key=' + text).then((res) => {
                        setData((res.data.data??[]).map((item) => {
                            return {
                                ...item,
                                Type: 'Liver',
                                Fans: item.ExtraInt
                            }
                        }))
                }
                )
            }
        }
    },[text,type])

    useEffect(() => {
        console.log(data);
    }, [data]);

    const redirect = useNavigate()

    useEffect(() => {
        var sorted = history.sort((a,b) => {b.Pinned-a.Pinned})
        if (sorted !== history) {
            setHistory(sorted)
        }
    }, [history]);
    useEffect(() => {
        const handleClick = () => {
            setOpenUser('')
        }
        document.addEventListener('click', handleClick)
        return () => {
            document.removeEventListener('click', handleClick)
        }
    }, [])
    const [rooms,setRooms] = React.useState(window.CACHED_ROOMS??[])
    useEffect(() => {
        axios.get("/api/hot").then((response) => {
            setRooms(response.data.data);
            window.CACHED_ROOMS = response.data.data;
        })
    }, []);
    const [openUser,setOpenUser] = React.useState('')
    return (
        <div className={'flex items-center justify-center flex-col'}>
            <div className={'w-full sm:w-[50vw] flex justify-center'}>
                <EyeIcon/>
            </div>
            <ScrollShadow className={'w-[80vw] sm:w-[50vw] max-h-[20vh] overflow-scroll scrollbar-hide mt-6'}>

                {rooms.map((room, index) => (
                    <NavLink to={'/liver/' + room.UserID} onMouseEnter={() => {
                        setAvatar(`${AVATAR_API}${room.UserID}`)
                    }}>
                        <Chip
                            avatar={<Avatar name={room.UserName} src={`${AVATAR_API}${room.UserID}`} />}
                            variant="flat"
                            className={'ml-2 mt-1'}

                        >
                            <p className={'font-bold'}>{room.UserName}</p>
                        </Chip>
                    </NavLink>))}
            </ScrollShadow>
            <div className={'w-full flex justify-center flex-col sm:flex-row'}>
                <Select className={'sm:max-w-xs'} label="Type" onChange={(e) => {
                    setType(e.target.value)
                }} defaultSelectedKeys={['Liver']} >
                    <SelectItem key={'Liver'}>Liver</SelectItem>
                    <SelectItem key={'Watcher'}>Watcher</SelectItem>
                </Select>
                <Autocomplete label={`UName or UID`} className={'ml-0 sm:ml-4 sm:mt-0 mt-2'} onInputChange={(e) => {
                    setText(e)
                    if (e.length >= 4 && text === '') {
                        if (isNumber(e) || isNumber(e.replace("UID:",""))) {
                            window.open(`${type === "Watcher" ? "/user/" : "/liver/"}${e.replace("UID:", "")}`)
                            axios.get("/api/user/card?uid=" + e.replace("UID:", "")).then((res) => {
                                var name = res.data.data.Name
                                setHistory(prevState => {
                                    return [{
                                        UID:e.replace("UID:", ""),
                                        UName:name,
                                        Pinned:false,
                                        Type:type,
                                    },...prevState.filter(item => item.UID !== e.replace("UID:",""))]
                                })
                            })
                        }
                    }
                }} inputValue={text} allowsCustomValue={true}>
                    {data.map((e) => {
                        return (
                            <AutocompleteItem key={String(e.UID)} textValue={`${e.UName} UID:${e.UID}`}  onClick={() => {
                                if (type === "Watcher") {
                                    window.open("/user/" + e.UID)
                                }
                                if (type === "Liver") {
                                    window.open("/liver/" + e.UID)
                                }
                                setHistory(prevState => {
                                    return [{
                                        UID:e.UID,
                                        UName:e.UName,
                                        Pinned:false,
                                        Type:type,
                                    },...prevState.filter(item => item.UID !== e.UID)]
                                })
                            }}>
                                <div className={'flex flex-row'}>
                                    <div className={'flex flex-col'}>
                                        <Avatar src={`${AVATAR_API}${e.UID}`} className={'mr-2'}/>

                                    </div>
                                    <div>
                                        <p className={'font-bold'}>{e.UName}</p>
                                        {type === "Liver" && <p className={' text-small'}>
                                            {e.Fans && e.Fans.toLocaleString()} Fans
                                        </p>}
                                        {type === "Watcher" && <p className={' text-small'}>
                                            <Chip
                                                startContent={e.MedalType ?<img src={getGuardIcon(e.MedalType)}/>:<CheckIcon size={18}/> }
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
                        )
                    })}
                </Autocomplete>
            </div>
            <div className={'w-full sm:w-[50vw] flex justify-center mt-4 flex-wrap select-none'}>
                {history.sort((a, b) => b.Pinned - a.Pinned).map((item, index) => (

                    <div ref={bind} data-uid={item.UID} data-pinned={item.Pinned} onContextMenu={() => {
                        setHistory(prevState => {
                            return prevState.map((item0,index) => {
                                if (item.UID === item0.UID) {
                                    return {
                                        ...item,
                                        Pinned:!item.Pinned
                                    }
                                }
                                return item0
                            })
                        })
                    }}               onClick={() => {
                        redirect(`/${item.Type === 'Liver' ? 'liver' : 'user'}/${item.UID}`)
                    }}>
                                <Chip
                                    avatar={<Avatar name={item.UName} src={`${AVATAR_API}${item.UID}`} />}
                                    variant="flat"
                                    className={'ml-2 mt-1'}
                                    color={item.Pinned ? 'success' : 'default'}
                                    onClose={(e) => {
                                        setHistory(history.filter(item0 => item0.UID !== item.UID))
                                    }}


                                >
                                    <p className={'font-bold'}>{item.UName}</p>
                                </Chip>
                    </div>

                    ))}
            </div>
        </div>
    );
}

export default IndexPage;
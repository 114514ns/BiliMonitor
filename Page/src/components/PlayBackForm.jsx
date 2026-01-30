import React, {useCallback, useEffect} from 'react';
import {
    Button, Listbox,
    ListboxItem,
    Modal,
    ModalBody,
    ModalContent,
    ModalFooter,
    ModalHeader,
    useDisclosure
} from "@heroui/react";
import {
    useRef,
    useState,
} from "react";
import axios from "axios";
import UserChip from "./UserChip";

function toBlob(text, mime = "text/plain;charset=utf-8") {
    const blob = new Blob([text], { type: mime });
    return URL.createObjectURL(blob);
}
function toBase64(str)
{
    const uint8Array = new TextEncoder().encode(str);
    return btoa(String.fromCharCode(...uint8Array));
}
function PlayBackForm(props) {

    const {isOpen, onOpen, onOpenChange} = useDisclosure();

    useEffect(() => {
        window.PLAY_BACK_OPEN = onOpen;
        console.log(props)

    },[])

    const hlsRef = React.useRef(null);


    const cacheMapping = React.useRef(new Map())
    const videoRef = useRef(HTMLVideoElement); //video元素ref

    const [link, setLink] = useState(
        ""
    ); //视频直链

    const [meta,setMeta] = useState({}); //元数据

    const [msg,setMsg] = useState([]) //所有弹幕

    const [off,setOff] = React.useState(0)

    const destroy = () => {
        console.log("destroy")
        hlsRef.current.destroy();
        hlsRef.current = null;
        videoRef.current.pause();
        videoRef.current.removeAttribute('src');
    }
    useEffect(() => {
        if (link !== "") {
            console.log("useEffect")
            var mapping = []

            async function handle() {
                if (link.includes("CONVERT.mp4") || link.includes("CONERT.mp4") || link.includes("COVERT")) {
                    videoRef.current.src = link
                    return;
                }
                if (cacheMapping.current.has(link.replace(".mp4", ".json")) ) {
                    mapping = cacheMapping.current.get(link.replace(".mp4", ".json"));
                } else {
                    mapping = await (await fetch(link.replace(".mp4", ".json"))).json()
                    cacheMapping.current.set(link.replace(".mp4", ".json"), mapping)
                }
                if (window.CACHE_REDIRECT === undefined) {
                    window.CACHE_REDIRECT = new Map()
                }
                var host = ""
                if (window.CACHE_REDIRECT.has(link)) {
                    host = window.CACHE_REDIRECT.get(link);
                } else {
                    host = (await (await axios.post("/api/redirect",{
                        data: toBase64(link)
                    })).data).url

                    window.CACHE_REDIRECT.set(link, host)
                }
                var dst = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-TARGETDURATION:4\n`;
                mapping.forEach((item) => {
                    dst = dst + "#EXTINF:4.000\n";
                    dst =
                        dst +
                        //host+
                        link +
                        `?start=${item.split(",")[0]}&end=${item.split(",")[1]}\n`;
                });
                dst = dst + '#EXT-X-ENDLIST\n'
                const hls = new Hls({
                    enableWorker: true,
                    lowLatencyMode: true,

                    fetchSetup: (context, init) => {
                        console.log(context);
                        if (context.url.includes("start=")) {
                            var u = URL.parse(context.url);
                            var range = `bytes=${u.searchParams.get(
                                "start"
                            )}-${u.searchParams.get("end")}`;
                            init.headers = {
                                ...init.headers,
                                Range: range,
                            };
                        }
                        return new Request(context.url , init);
                    },

                });

                hls.on(Hls.Events.ERROR, (event, data) => {
                    console.error("HLS Error:", data);
                    if (data.fatal) {
                        switch (data.type) {
                            case Hls.ErrorTypes.NETWORK_ERROR:
                                console.error("Network error");
                                hls.startLoad();
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                console.error("Media error");
                                hls.recoverMediaError();
                                break;
                            default:
                                hls.destroy();
                                break;
                        }
                    }
                });

                hls.on(Hls.Events.MEDIA_ATTACHED, () => {
                    console.log("Media attached, loading source");
                    hls.loadSource(toBlob(dst))
                });

                hlsRef.current = hls;
            }
            handle()
        }
        return () => {
            if (hlsRef.current) {
                destroy()
                //setLink('')
            }

        };
    }, [link]);

    useEffect(() => {
        if (off > 0) {
            setTimeout(() => {
                videoRef.current.currentTime = off-3;
            },1000)
        }
    },[off])

    useEffect(() => {
        props.items.length &&     fetch(props.items.filter((s) => s.Link.includes("metadata.json"))[0].Link).then(async (res) => {
            setMeta(await res.json()??{})
        })
    },[props.items])

    useEffect(() => {
        if (props.id > 0) {
            axios.get(`/api/live/${props.id}/?page=1&limit=5000&order=undefined&mid=0`).then( (res) => {
                setMsg(res.data.records)
            })
            cacheMapping.current = new Map()
        }
    },[props.id])




    return (
        <div className={'w-[80vw]'}>
            <Modal isOpen={isOpen} onOpenChange={onOpenChange} size={'full'} scrollBehavior={'inside'}>
                <ModalContent>
                    {(onClose) => (
                        <>
                            <ModalHeader className="flex flex-col gap-1">Playback</ModalHeader>
                            <ModalBody>
                                <div className={'flex lg:flex-row flex-col'}>
                                    <Listbox aria-label="Actions" onAction={(key)  => {
                                        console.log(key);
                                        setLink(key)
                                    }} className={'w-[180px]'}>
                                        {props.items.filter((e) => {
                                            return e.FileName.includes(".mp4")
                                        }).map((i,index) => {
                                            var color = i.Link === link?'#E1F5FF':''
                                            return (
                                                <ListboxItem
                                                    key={i.Link}
                                                    className=""
                                                    style={{backgroundColor:color}}
                                                    onPress={() => {

                                                    }}
                                                >
                                                    <span className="text-sm">{i.FileName}</span>
                                                    {meta.ChunkRecord && meta.ChunkRecord[index] && <p className={'font-light'}>{new Date(meta.ChunkRecord[index]).toLocaleTimeString()}</p>}
                                                </ListboxItem>
                                            )
                                        })}
                                    </Listbox>
                                    <div className={'flex flex-row'}>
                                        <video ref={videoRef} controls autoPlay playsInline
                                               className={'w-[100%] lg:w-[84%] h-auto'}></video>
                                        <div className={'flex flex-col overflow-scroll max-h-[85vh] ml-4'}>
                                            {msg.map((e,index) => {
                                                return (
                                                    <div>
                                                        {(index === 0?true:(e.FromId !== msg[index-1].FromId) )&& <UserChip props={e}/>}
                                                        <span
                                                            className={'hover:text-[#0AA5D8]'}
                                                            onClick={() => {
                                                                var index = 0
                                                                if (meta && meta.ChunkRecord) {
                                                                    meta.ChunkRecord.forEach((chunk,i) => {
                                                                        if (new Date(chunk).getTime()+3600*8 < new Date(e.CreatedAt).getTime()) {
                                                                            index = i
                                                                        }
                                                                    })
                                                                    setOff( (new Date(e.CreatedAt).getTime() - new Date(meta.ChunkRecord[index]).getTime() + 3600*8)/1000)
                                                                    setLink(props.items.filter((e) => {
                                                                        return e.FileName.includes(".mp4")
                                                                    })[index].Link)
                                                                    console.log(off)
                                                                    console.log(index)
                                                                }
                                                            }}
                                                        >{e.Extra}{e.GiftName &&   e.GiftName + '* ' + e.GiftAmount.Int16 +'  CNY ' +e.GiftPrice.Float64  }</span>
                                                    </div>

                                                )
                                            })}
                                        </div>
                                    </div>
                                    <div>

                                    </div>
                                </div>
                            </ModalBody>
                            <ModalFooter>
                                <Button color="danger" variant="light" onPress={() => {
                                    onClose()
                                    destroy()
                                }}>
                                    Close
                                </Button>
                                <Button color="primary" onPress={() => {
                                    onClose()
                                    destroy()
                                }}>
                                    Action
                                </Button>
                            </ModalFooter>
                        </>
                    )}
                </ModalContent>
            </Modal>
        </div>
    );
}

export default PlayBackForm;

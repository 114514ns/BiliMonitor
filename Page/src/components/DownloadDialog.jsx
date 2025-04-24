import React, {useState} from 'react';
import {
    addToast,
    Button,
    Checkbox,
    CheckboxGroup,
    Input,
    Modal,
    ModalBody,
    ModalContent,
    ModalFooter,
    ModalHeader,
    ToastProvider
} from "@heroui/react";
import PubSub from "pubsub-js";
import axios from "axios";
import LiveCard from "./LiveCard";

function DownloadDialog(props) {
    const {isOpen, setOpen} = useState(props.isOpen);
    const [video, setVideo] = useState([]);
    const host = location.hostname;


    const port =location.port
    const [id, setId] = useState("");
    const protocol = location.protocol.replace(":", "")
    const [showCard, setShowCard] = useState(false);
    const [selectedVideo, setSelectedVideo] = useState([]);
    var real = ""
    var type = "video"
    return (
        <>
            <ToastProvider placement={'top-center'} toastOffset={60}/>
            <Modal isOpen={props.isOpen} onClose={() => {
                PubSub.publish('DownloadDialog', 'Close')
            }} size="lg" style={{
                maxHeight: '80vh',
                overflowY: 'auto',
            }}>
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">下载视频</ModalHeader>
                    <ModalBody>
                        <Input label={'BV号/视频链接/播放列表链接'} onChange={(e) => {
                            setId(e.target.value)
                        }}></Input>
                        <Button onPress={() => {

                            try {
                                var path = new URL(id).pathname.split("/")
                                if (id.indexOf("video/BV") !== -1) {
                                    real = path[2]
                                }
                                if (id.indexOf("lists") !== -1) {
                                    type = 'list'
                                    axios.get(`${protocol}://${host}:${port}/api/parseList?mid=${path[1]}&season=${path[3]}`).then((res) => {
                                        setVideo(res.data.data)
                                        setShowCard(true)
                                    })
                                }
                            } catch (e) {
                                if (id.indexOf("BV") !== -1) {
                                    real = id
                                }

                            }
                            if (type === "video") {
                                axios.get(`${protocol}://${host}:${port}/api/parse?bv=${real}`).then(res => {
                                    setVideo(res.data.data)
                                    setShowCard(true)
                                })
                            }
                        }}>提交</Button>
                        {showCard ? <div
                            style={{width: '100%', display: 'flex', justifyContent: 'center', alignItems: 'center'}}>
                            <div>
                                <CheckboxGroup onChange={e => {
                                    setSelectedVideo(e)
                                }}>

                                    {video.map(element => {
                                        return (
                                            <Checkbox value={element}>
                                                <LiveCard liveData={{
                                                    Live: null,
                                                    UName: element.Author,
                                                    UID: element.UID,
                                                    Title: element.Title,
                                                    Cover: element.Img,
                                                    Cid:element.Cid,
                                                    Face: element.AuthorFace,
                                                    Area: ""
                                                }}></LiveCard>
                                            </Checkbox>

                                        )
                                    })}
                                </CheckboxGroup>
                            </div>
                        </div> : <div></div>}
                    </ModalBody>
                    <ModalFooter>
                        <Button color="danger" variant="light" onPress={() => {
                            PubSub.publish('DownloadDialog', 'Close')
                        }}>
                            Close
                        </Button>
                        <Button color="primary" onPress={() => {
                            PubSub.publish('DownloadDialog', 'Close')
                            selectedVideo.forEach(video => {
                                axios.get(`${protocol}://${host}:${port}/api/download?bv=${video.BV}&part=${video.Part}`)
                            })
                            addToast({
                                title: "添加成功",
                                description: "请稍后自行前往Alist查看",
                                color: "success",
                            })
                        }}>
                            Action
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </>
    );
}

export default DownloadDialog;
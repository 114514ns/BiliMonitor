import React, {useEffect, useRef, useState} from 'react';
import axios from "axios";
import {video} from "framer-motion/m";

function BVPlayer(props) {

    const [v,setV] = useState("");

    const [a,setA] = useState("");

    const videoRef = useRef(HTMLVideoElement);

    const audioRef = useRef(HTMLAudioElement);

    useEffect(() => {
        if (props.bv !== '') {
            axios.get("/api/bv/view?bv=" + props.bv).then((res) => {
                var dash = res.data.data.dash

                setV(`https://stream-proxy.ikun.dev?url=${btoa(dash.video[0].base_url)}`)
                setA(`https://stream-proxy.ikun.dev?url=${btoa(dash.audio[0].base_url)}`)
            })
        }

    },[props.bv])

    useEffect(() => {
        if (videoRef.current && audioRef.current) {
            const video = videoRef.current;
            const audio = audioRef.current;
            video.addEventListener('play', () => {
                audio.play();
            });
            video.addEventListener('pause', () => {
                audio.pause();
            });

            video.addEventListener('waiting', () => {
                audio.pause();
            });

            video.addEventListener('playing', () => {
                if (!video.paused) {
                    audio.play();
                }
            });
            video.addEventListener('seeked', () => {
                audio.currentTime = video.currentTime;
                if (!video.paused) {
                    if (video.readyState >= 3) {
                        audio.play();
                    }
                }
            });
        }
    },[])
    return (
        <div>
            <video ref={videoRef} src={v} controls/>

            <audio ref={audioRef} src={a} className={'hidden'}/>
        </div>
    )
}

export default BVPlayer;
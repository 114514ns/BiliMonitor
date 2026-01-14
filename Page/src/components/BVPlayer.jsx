import React, {forwardRef, useEffect, useRef, useState} from 'react';
import axios from "axios";



const BVPlayer = forwardRef(function BVPlayer(props, ref) {
    const [v, setV] = useState("");
    const [a, setA] = useState("");

    const videoRef = useRef(null);
    const audioRef = useRef(null);

    useEffect(() => {
        if (props.bv !== '') {
            const sp = props.bv.split("-");
            const b = sp[0];
            const c = sp.length === 2 ? sp[1] : '';

            axios
                .get(`/api/bv/view?bv=${b}&cid=${c}`)
                .then((res) => {
                    const dash = res.data.data.dash;
                    setV(`https://stream-proxy.ikun.dev?url=${btoa(dash.video[0].base_url)}`);
                    setA(`https://stream-proxy.ikun.dev?url=${btoa(dash.audio[0].base_url)}`);
                });
        }
    }, [props.bv]);

    useEffect(() => {
        if (!videoRef.current || !audioRef.current) return;

        const video = videoRef.current;
        const audio = audioRef.current;

        const onPlay = () => audio.play();
        const onPause = () => audio.pause();
        const onWaiting = () => audio.pause();
        const onPlaying = () => {
            if (!video.paused) audio.play();
        };
        const onSeeked = () => {
            audio.currentTime = video.currentTime;
            if (!video.paused && video.readyState >= 3) {
                audio.play();
            }
        };

        video.addEventListener('play', onPlay);
        video.addEventListener('pause', onPause);
        video.addEventListener('waiting', onWaiting);
        video.addEventListener('playing', onPlaying);
        video.addEventListener('seeked', onSeeked);

        return () => {
            video.removeEventListener('play', onPlay);
            video.removeEventListener('pause', onPause);
            video.removeEventListener('waiting', onWaiting);
            video.removeEventListener('playing', onPlaying);
            video.removeEventListener('seeked', onSeeked);
        };
    }, []);

    return (
        <div ref={ref} className={props.className}>
            <video
                ref={videoRef}
                src={v}
                controls
                className="w-full"
            />
            <audio
                ref={audioRef}
                src={a}
                className="hidden"
            />
        </div>
    );
});

export default BVPlayer;
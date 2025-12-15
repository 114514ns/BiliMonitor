import React, {useEffect} from 'react';
import {
    Modal,
    ModalBody,
    ModalContent,
    ModalFooter,
    ModalHeader,
    Slider,
    Image,
    Select,
    SelectItem
} from "@heroui/react";
import axios from "axios";

function SettingDialog(props) {

    const [opacity,setOpacity] = React.useState(getOpacity)
    useEffect(() => {
        localStorage.setItem("opacity",opacity)
    },[opacity])


    const images   = isMobile()? [
        'https://storage.ikun.dev/d/Microsoft365/static/136455890_p0.avif?sign=w3tPqeFMgX9NIwLn4mq8FI9XK8pgbcf68Ir9crpD9QQ=:0' ,//栞栞shiori
        'https://storage.ikun.dev/d/Microsoft365/static/133487584_p0_master1200.avif?sign=gqjurkYhTQwdT_i2Z-MfmrDKXOc3INkqlUYQj1L--Vg=:0', //犬戎
        'https://storage.ikun.dev/d/Microsoft365/static/83307417_p0.avif?sign=wxmZ7R-4WVEy1ils1yK5fyeLYEFlSEwJZvI_luSoL3s=:0', //白神遥Haruka
    ]:[

        //'https://storage.ikun.dev/d/Microsoft365/static/109780395_p0.avif?sign=fagwpdupRN5rGXmaqNIPLByxNPT08KFJ6XHcLYFVcG4=:0', //嘉然
        'https://storage.ikun.dev/d/Microsoft365/static/136963002_p8.avif?sign=7lxaLKnSTkTgi8PYe5wnTgxXHqYx89jW6yhL9WGVblI=:0', //折原鹿鹿
        'https://storage.ikun.dev/d/Microsoft365/static/videoframe_3241.avif?sign=HleQqGXynCxyCyAe4yQM1e6G9X0JvGFjO0habFV2GLA=:0', //KUN
    ]
    const ref = React.useRef()
    return (
        <div className={'max-h-1/2'}>
            <Modal onClose={props.onClose} isOpen={true} size={'lg'} className={'max-h-1/2 sm:max-h-2/3'} scrollBehavior={'inside'}>
                <ModalContent>
                    <ModalHeader>
                        Settings
                    </ModalHeader>
                    <ModalBody>
                        <Slider
                            className="max-w-md"
                            defaultValue={opacity}
                            label="Opacity"
                            maxValue={100}
                            minValue={0}
                            step={1}
                            onChange={(e) => {
                                props.onOpacityChange(e)
                                e && setOpacity(e)
                                e[0] && setOpacity(e[0])
                            }}

                        />

                        <p2>
                            Images
                        </p2>

                        <div>
                            <picture
                                className="flex items-center justify-center"
                                style={{
                                    width: ref.current?.width ?? 0,
                                    height: ref.current?.height ?? 0
                                }}
                                onClick={() => {
                                    var html = document.querySelector('html')
                                    html.style.backgroundImage = ``
                                    props.onOpacityChange(100)
                                    localStorage.setItem("background",'')
                                    setOpacity(100)
                                }}
                            >
                                <img src={'data:image/svg+xml;utf8,%3Csvg viewBox=\'0 0 24 24\' display=\'inline-block\' vertical-align=\'middle\' width=\'1.2em\' height=\'1.2em\' xmlns=\'http://www.w3.org/2000/svg\' %3E%3Cg fill=\'none\' stroke=\'currentColor\' stroke-linecap=\'round\' stroke-linejoin=\'round\' stroke-width=\'2\'%3E%3Cpath d=\'M15 8h.01M7 3h11a3 3 0 0 1 3 3v11m-.856 3.099A3 3 0 0 1 18 21H6a3 3 0 0 1-3-3V6c0-.845.349-1.608.91-2.153\'/%3E%3Cpath d=\'m3 16l5-5c.928-.893 2.072-.893 3 0l5 5m.33-3.662c.574-.054 1.155.166 1.67.662l3 3M3 3l18 18\'/%3E%3C/g%3E%3C/svg%3E'}/>

                            </picture>
                            {images.map((item) => {
                                return (
                                    <Image src={item} className={'mt-2'} isZoomed onClick={(e) => {
                                        var html = document.querySelector('html')
                                        html.style.backgroundImage = `url("${item}")`
                                        html.style.backgroundSize = 'cover'
                                        html.style.backgroundAttachment = 'fixed'
                                        html.style.backgroundPosition = 'center'
                                        localStorage.setItem("background",item)
                                        props.onOpacityChange(80)
                                        setOpacity(80)
                                        axios.get(item + "@.avg_color").then((res) => {
                                            
                                        })
                                    }} ref={ref}/>
                                )
                            })}
                        </div>

                        <p>Default Page Size</p>

                        <Select className=" mt-4 mb-4" label={'Page Size'} defaultSelectedKeys={[localStorage.getItem("defaultPageSize")??'10']}>
                            <SelectItem onClick={e => { localStorage.setItem("defaultPageSize","10") }} key={'10'}>
                                10
                            </SelectItem>
                            <SelectItem onClick={e => { localStorage.setItem("defaultPageSize","50")}} key={'50'}>
                                50
                            </SelectItem>
                            <SelectItem onClick={e => { localStorage.setItem("defaultPageSize","200") }} key={'200'}>
                                200
                            </SelectItem>
                            <SelectItem onClick={e => { localStorage.setItem("defaultPageSize","500")}} key={'500'}>
                                500
                            </SelectItem>
                        </Select>



                    </ModalBody>
                    <ModalFooter>

                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div>
    );
}

function generateTextColorFromBackground (backgroundColor) {
    // 支持 #RGB / #RRGGBB / rgb(r,g,b)
    const parseColor = (str) => {
        str = str.trim();
        if (str.startsWith('#')) {
            if (str.length === 4) {
                const r = parseInt(str[1] + str[1], 16);
                const g = parseInt(str[2] + str[2], 16);
                const b = parseInt(str[3] + str[3], 16);
                return [r, g, b];
            } else if (str.length === 7) {
                const r = parseInt(str.slice(1, 3), 16);
                const g = parseInt(str.slice(3, 5), 16);
                const b = parseInt(str.slice(5, 7), 16);
                return [r, g, b];
            }
        } else if (str.startsWith('rgb')) {
            const nums = str.match(/\d+/g)?.map(Number);
            if (nums && nums.length >= 3) return nums.slice(0, 3);
        }
        throw new Error('Unsupported color format: ' + str);
    };

    // 相对亮度 (W3C公式)
    const relativeLuminance = (r, g, b) => {
        const srgb = [r, g, b].map(v => {
            v /= 255;
            return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
        });
        return 0.2126 * srgb[0] + 0.7152 * srgb[1] + 0.0722 * srgb[2];
    };

    // 计算对比度比值
    const contrastRatio = (lum1, lum2) => {
        const L1 = Math.max(lum1, lum2);
        const L2 = Math.min(lum1, lum2);
        return (L1 + 0.05) / (L2 + 0.05);
    };

    const [r, g, b] = parseColor(backgroundColor);
    const bgLum = relativeLuminance(r, g, b);

    const whiteLum = relativeLuminance(255, 255, 255);
    const blackLum = relativeLuminance(0, 0, 0);

    const contrastWithWhite = contrastRatio(bgLum, whiteLum);
    const contrastWithBlack = contrastRatio(bgLum, blackLum);

    // 如果黑白都不够，就生成增强色
    if (Math.max(contrastWithWhite, contrastWithBlack) < 4.5) {
        // 根据背景色选择增强颜色
        const enhancedColor =
            bgLum > 0.5
                ? 'rgb(30, 30, 60)' // 背景亮 → 深蓝灰文字
                : 'rgb(240, 230, 210)'; // 背景暗 → 柔和亮米色文字
        return enhancedColor;
    }

    // 否则返回对比度更高的那个
    return contrastWithWhite > contrastWithBlack ? '#ffffff' : '#000000';
}
export default SettingDialog;
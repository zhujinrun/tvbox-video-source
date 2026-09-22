/**
 * Created by DreamBoy on 2023/3/8.
 */
$(function () {
    $.fn.initInputGroup = function (options) {
        //1.Settings 初始化设置
        var c = $.extend({
            'widget': 'input',
            'add': "<span class=\"glyphicon glyphicon-plus\"></span>",
            'del': "<span class=\"glyphicon glyphicon-minus\"></span>",
            'key': "tv"
        }, options);

        var _this = $(this);

        if (c.data) {
            var length = Object.keys(c.data).length, index = 0;
            // console.log(count);
            for (let key in c.data) {
                //console.log(key, c.data[key])
                //添加序号为`key`的数字下标的输入框组
                addInputGroup(key.replace(c.key, ''), c.data[key], index >= length - 1);
                index++;
            }
            if (index == 0) {
                //添加序号为1的输入框组
                addInputGroup(1);
            }
        } else {
            console.log('初始化...');
            //添加序号为1的输入框组
            addInputGroup(1);
        }

        /**
         * 添加序号为order的输入框组
         * @param order 输入框组的序号
         */
        function addInputGroup(order, val = '', add = true) {
            val = (val == null ? "" : val);
            if ("vip" == order) {
                $(".input-group:first").children('.form-control').val(val);
            } else {
                //1.创建输入框组
                var inputGroup = $("<div class='input-group'></div>");
                //2.输入框组的序号
                var inputGroupAddon1 = $("<span class='input-group-addon'></span>");
                //3.设置输入框组的序号
                inputGroupAddon1.html(c.key + order + "");

                //4.创建输入框组中的输入控件（input或textarea）
                var widget = '', inputGroupAddon2;
                if (c.widget == 'textarea') {
                    widget = $("<textarea class='form-control' style='resize: vertical;'>" + val + "</textarea>");
                    inputGroupAddon2 = $("<span class='input-group-addon'></span>");
                } else if (c.widget == 'input') {
                    widget = $("<input class='form-control' type='text' value='" + val + "'/>");
                    inputGroupAddon2 = $("<span class='input-group-btn'></span>");
                }

                //5.创建输入框组中最后面的操作按钮
                var addBtn = add ? $("<button class='btn btn-default' type='button'>" + c.add + "</button>") : $("<button class='btn btn-default' type='button'>" + c.del + "</button>");
                addBtn.appendTo(inputGroupAddon2).on('click', function () {
                    //6.响应删除和添加操作按钮事件
                    if ($(this).html() == c.del) {
                        $(this).parents('.input-group').remove();
                    } else if ($(this).html() == c.add) {
                        $(this).html(c.del);
                        addInputGroup(order + 1);
                    }
                    //7.重新排序输入框组的序号
                    resort();
                });

                inputGroup.append(inputGroupAddon1).append(widget).append(inputGroupAddon2);

                _this.append(inputGroup);
            }
        }

        function resort() {
            var child = _this.children();
            $.each(child, function (i) {
                $(this).find(".input-group-addon").eq(0).html(c.key + (i + 1) + '');
            });
        }
    }
});
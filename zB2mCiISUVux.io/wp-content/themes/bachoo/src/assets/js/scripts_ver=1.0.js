$(document).ready(function () {
	// Passive event listeners
	jQuery.event.special.touchstart = {
		setup: function( _, ns, handle ) {
			this.addEventListener("touchstart", handle, { passive: !ns.includes("noPreventDefault") });
		}
	};
	jQuery.event.special.touchmove = {
		setup: function( _, ns, handle ) {
			this.addEventListener("touchmove", handle, { passive: !ns.includes("noPreventDefault") });
		}
	};
	// IMPORTANT to Remove animate Once every time page is loaded for one time animation
    sessionStorage.removeItem('animate_once_reveal_scale');
    sessionStorage.removeItem('animate_once_reveal_split');
    sessionStorage.removeItem('animate_once_reveal_text_left');
    sessionStorage.removeItem('animate_once_reveal_top');
    sessionStorage.removeItem('animate_once_reveal_simple');

    var data_prices_yearly = [];
    var data_prices_monthly = [];
    var data_prices_quarterly = []
    try {

        function checkemailvalidation(email) {
            var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
            return re.test(String(email).toLowerCase());
        }
        var authWindow;
        var stripe = Stripe(stripe_publish_key)
        var isAjax = false;
        $('body').on('click', '.SignUpBtn', function (e) {

            // Prevent the form from getting submitted
            e.preventDefault();

            var $this = $(this);
            $('.sign-up__form input.textBox').each(function () {

                if ($(this).val() === '') {
                    if ($(this).hasClass('required')) {
                        $(this).addClass('warning');
                        $(this).parent().addClass('warning');

                        $(this).parent().find('.input-error').show();
                        $(this).parent().find('.input-error').html('This field is required should not be empty.');
                    }
                }

            });

            if (!$('.sign-up__agreement_checkbox').is(':checked')) {
                $('.sign-up__agreement_checkbox').parent().find('.input-error').show()
                $('.sign-up__agreement_checkbox').parent().find('.input-error').html('Please tick the box before proceeding.');
                $('.sign-up__agreement_checkbox').parent().addClass('warning')
            } else {
                $('.sign-up__agreement_checkbox').parent().find('.input-error').hide()
                $('.sign-up__agreement_checkbox').parent().removeClass('warning')
            }
            if (checkemailvalidation($('.email').val()) === false) {

                $('.email').parent().find('.input-error').show();
                $('.email').parent().find('.input-error').html('Please enter a valid email address');
                $('.email').parent().addClass('warning');
                $('.email').addClass('warning');

            }

            if ($('.password').val() !== $('.confirmPassword').val()) {

                $('.confirmPassword').parent().find('.input-error').show();
                $('.confirmPassword').parent().find('.input-error').html('Password does not match');
                $('.confirmPassword').addClass('warning');
                $('.confirmPassword').parent().addClass('warning');
                $('.password').addClass('warning');

            }


            if ($('.fname').val() !== '' && checkemailvalidation($('.email').val()) === true) {
                if ($('.password').val() !== '' && $('.confirmPassword').val() !== '') {
                    if ($('.password').val() === $('.confirmPassword').val() && $('.sign-up__agreement_checkbox').is(':checked')) {

                        $('.sign-up__form .noticeBox').hide();
                        $('.sign-up__form input.textBox').removeClass('warning');
                        $('.sign-up__form label .input-error').hide().html('');
                        // Stripe.setPublishableKey(stripe_publish_key)
                        // alert('Send ajax request');
                        if (!$this.hasClass('processing')) {
                            $this.addClass('processing').attr('disabled', true);
                            $this.find('span').hide();
                            $this.find('#dvLoading').show();
                            StripeCheckout.open({
                                key: stripe_publish_key,
                                billingAddress: true,
                                currency: 'usd',
                                name: 'Registration',
                                image: site_url + '/wp-content/uploads/2022/04/bb-GREEN.png',
                                allowRememberMe: false,
                                email: jQuery('input[type=email]').val(),
                                panelLabel: 'Proceed',
                                locale: 'auto',
                                token: token,
                                opened: function () {

                                },
                                closed: function () {
                                    if (!isAjax) {
                                        $('.SignUpBtn').removeClass('processing').attr('disabled', false);
                                        $this.find('span').show();
                                        $this.find('#dvLoading').hide();
                                    }
                                }
                            })

                        }

                    }
                }
            }

        });
        $('body').on('click', '.plans__info.from-thankyou .button', function(e){
            e.preventDefault();
            $this = $(this);
            $this.find('span').hide();
            var plan_type = $($this.parent()).find('h4').text() == 'Starter' ? 'Starter' : $($this.parent()).find('h4').text().toLowerCase();  ;
            localStorage.setItem('ty_plan_type', plan_type);
            if(!$this.hasClass('processing')){
                $('.plans__list .button').addClass('processing').attr('disabled', true);
                $this.find('.sk-circle').show();

                StripeCheckout.open({
                    key: stripe_publish_key,
                    billingAddress: true,
                    currency: 'usd',
                    name: 'BookBeam Upgrade',
                    image: site_url + '/wp-content/uploads/2022/04/bb-GREEN.png',
                    allowRememberMe: false,
                    panelLabel: 'Upgrade',
                    email: jQuery('input[type=hidden].hidden__email').val(),
                    locale: 'auto',
                    token: thankyou_token,
                    opened: function () {

                    },
                    closed: function () {
                        if (!isAjax) {
                            $('.plans__list .button').removeClass('processing').attr('disabled', false);
                            $this.find('span').show();
                            $this.find('.sk-circle').hide();
                        }
                    }
                })
            }
        })
        $('body').on('click', '.plans__features.from-thankyou .button', function(e){
            e.preventDefault();
            $this = $(this);
            $this.find('span').hide();
            var plan_title = $(this).parent().parent().find('.plans__info h4').text();
        
            var plan_type = plan_title == 'Starter' ? 'Starter' : plan_title.toLowerCase();  ;
            localStorage.setItem('ty_plan_type', plan_type);
            if(!$this.hasClass('processing')){
                $('.plans__list .button').addClass('processing').attr('disabled', true);
                $this.find('.sk-circle').show();
                $('.errorNotice').removeClass('show')
                StripeCheckout.open({
                    key: stripe_publish_key,
                    billingAddress: true,
                    currency: 'usd',
                    name: 'BookBeam Upgrade',
                    image: site_url + '/wp-content/uploads/2022/04/bb-GREEN.png',
                    allowRememberMe: false,
                    panelLabel: 'Upgrade',
                    email: jQuery('input[type=hidden].hidden__email').val(),
                    locale: 'auto',
                    token: thankyou_token,
                    opened: function () {

                    },
                    closed: function () {
                        if (!isAjax) {
                            $('.plans__list .button').removeClass('processing').attr('disabled', false);
                            $this.find('span').show();
                            $this.find('.sk-circle').hide();
                        }
                    }
                })
            }
        })
        
        $('body').on('click', '.BFSignUpBtn', function(e){
            // Prevent the form from getting submitted
            e.preventDefault();

            var $this = $(this);
            $('.sign-up__form input.textBox').each(function () {

                if ($(this).val() === '') {
                    if ($(this).hasClass('required')) {
                        $(this).addClass('warning');
                        $(this).parent().addClass('warning');

                        $(this).parent().find('.input-error').show();
                        $(this).parent().find('.input-error').html('This field is required should not be empty.');
                    }
                }

            });

            if (!$('.sign-up__agreement_checkbox').is(':checked')) {
                $('.sign-up__agreement_checkbox').parent().find('.input-error').show()
                $('.sign-up__agreement_checkbox').parent().find('.input-error').html('Please tick the box before proceeding.');
                $('.sign-up__agreement_checkbox').parent().addClass('warning')
            } else {
                $('.sign-up__agreement_checkbox').parent().find('.input-error').hide()
                $('.sign-up__agreement_checkbox').parent().removeClass('warning')
            }
            if (checkemailvalidation($('.email').val()) === false) {

                $('.email').parent().find('.input-error').show();
                $('.email').parent().find('.input-error').html('Please enter a valid email address');
                $('.email').parent().addClass('warning');
                $('.email').addClass('warning');

            }

            if ($('.password').val() !== $('.confirmPassword').val()) {

                $('.confirmPassword').parent().find('.input-error').show();
                $('.confirmPassword').parent().find('.input-error').html('Password does not match');
                $('.confirmPassword').addClass('warning');
                $('.confirmPassword').parent().addClass('warning');
                $('.password').addClass('warning');

            }


            if ($('.fname').val() !== '' && checkemailvalidation($('.email').val()) === true) {
                if ($('.password').val() !== '' && $('.confirmPassword').val() !== '') {
                    if ($('.password').val() === $('.confirmPassword').val() && $('.sign-up__agreement_checkbox').is(':checked')) {

                        $('.sign-up__form .noticeBox').hide();
                        $('.sign-up__form input.textBox').removeClass('warning');
                        $('.sign-up__form label .input-error').hide().html('');
                        if (!$this.hasClass('processing')) {
                            $this.addClass('processing').attr('disabled', true);
                            $this.find('span').hide();
                            $this.find('#dvLoading').show();
                            StripeCheckout.open({
                                key: stripe_publish_key,
                                billingAddress: true,
                                currency: 'usd',
                                name: 'Registration',
                                image: site_url + '/wp-content/uploads/2022/04/bb-GREEN.png',
                                allowRememberMe: false,
                                email: jQuery('input[type=email]').val(),
                                panelLabel: 'Proceed',
                                locale: 'auto',
                                token: bf_token,
                                opened: function () {

                                },
                                closed: function () {
                                    if (!isAjax) {
                                        $('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
                                        $this.find('span').show();
                                        $this.find('#dvLoading').hide();
                                    }
                                }
                            })

                        }
                    }
                }
            }
        })
        function thankyou_token(res){
            
            if(res.error){
                $('.plans__info .button, .plans__features .button').removeClass('processing')
                $('.plans__info .button, .plans__features .button').find('span').show();
                $('.plans__info .button, .plans__features .button').find('.sk-circle').hide();
                $('.plans__info .button, .plans__features .button').attr('disabled', false);
                return
            }
            if(res.id){
                isAjax = true;
                var plan_type = localStorage.getItem('ty_plan_type');
                var stripeToken = res.id;
                var plan_period = $('.pricing__switch').attr('data-pay-switch') == 'monthly' ? $('.pricing__switch').attr('data-pay-switch') : 'annual';
                
                var data = {
                    'action': 'bookbeam_change_plan',
                    'nonce': book_beam_params.nonce,
                    'plan_type': plan_type,
                    'plan_period': plan_period,
                    'stripeToken': stripeToken,
                    'email': jQuery('input[type=hidden].hidden__email').val(),
                    'applied_coupon': jQuery('input.coupon__input').val(),
                }
                $.ajax({
                    url: book_beam_params.ajax_url,
                    type: 'POST',
                    data: data,
                    beforeSend: function(e){
                        localStorage.removeItem('ty_plan_type');
                    },  
                    success: function (response) {
                        if(response.success){
                            if(response.status == 'pending_for_auth'){
                                stripe.confirmCardPayment(response.pi_client)
                                .then(result => {
                                    try{
                                        if(result.error){
                                            $('.plans__info .button, .plans__features .button').removeClass('processing')
                                            $('.plans__info .button, .plans__features .button').find('span').show();
                                            $('.plans__info .button, .plans__features .button').find('.sk-circle').hide();
                                            $('.plans__info .button, .plans__features .button').attr('disabled', false);
                                            return;
                                        }
                                    }catch(e){}
                                    if(result.paymentIntent.status == 'succeeded'){
                                        //Add ajax for inserting data to both app and site
                                        $.ajax({
                                            url: book_beam_params.ajax_url,
                                            type: 'POST',
                                            data: {
                                                action: 'bookbeam_change_plan_after_3ds',
                                                plan_type: plan_type,
                                                nonce: book_beam_params.nonce,
                                                email: jQuery('input[type=hidden].hidden__email').val(),
                                                applied_coupon: jQuery('input.coupon__input').val(),
                                                stripeToken: stripeToken,
                                                plan_period: plan_period,
                                                subs_id: response.subs_id,
                                                customer_id : response.customer_id
                                            },
                                            success: function (resp) {
    
                                                if (resp.success) {
                                                    window.dataLayer.push({
                                                        'event' : 'SignUpSubmit',
                                                        'eventCategory' : 'Sign up',
                                                        'eventAction' : 'Form Submit'
                                                      });
                                                    window.location.href = resp.redirect
                                                }
    
                                                if (!resp.success) {
                                                    $('.plans__info .button, .plans__features .button').removeClass('processing')
                                                    $('.plans__info .button, .plans__features .button').find('span').show();
                                                    $('.plans__info .button, .plans__features .button').find('.sk-circle').hide();
                                                    $('.plans__info .button, .plans__features .button').attr('disabled', false);
                                                    $('.errorNotice p').text(response.message);
                                                    $('.errorNotice').addClass('show')
                                                }
                                            }
                                        })
                                    }else{
                                        $('.plans__info .button, .plans__features .button').removeClass('processing')
                                        $('.plans__info .button, .plans__features .button').find('span').show();
                                        $('.plans__info .button, .plans__features .button').find('.sk-circle').hide();
                                        $('.plans__info .button, .plans__features .button').attr('disabled', false);
                                        $('.errorNotice p').text(response.message);
                                        $('.errorNotice').addClass('show')
                                    }
                                }).catch(err => {
                                    console.log({siErr: err});
                                });
                            }else{
                                window.dataLayer.push({
                                    'event' : 'SignUpSubmit',
                                    'eventCategory' : 'Sign up',
                                    'eventAction' : 'Form Submit'
                                });
                                window.location.href = response.redirect;
                            }
                        }else{
                            $('.plans__info .button, .plans__features .button').removeClass('processing')
                            $('.plans__info .button, .plans__features .button').find('span').show();
                            $('.plans__info .button, .plans__features .button').find('.sk-circle').hide();
                            $('.plans__info .button, .plans__features .button').attr('disabled', false);
                            $('.errorNotice p').text(response.message);
                            $('.errorNotice').addClass('show')
                        }
                    }
                })

            }
        }
        function token(res) {

            var $form = jQuery('.sign-up__form');
            // show processing message, disable links and buttons until form is submitted and reloads
            jQuery('a').bind("click", function () { return false; });
            jQuery('button').prop('disabled', true);
            jQuery('.overlay-container').show();

            if (res.error) {
                jQuery('.overlay-container').hide();
                jQuery('a').unbind("click");
                jQuery('button').prop('disabled', false);
                jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                jQuery('.sign-up__form .noticeBox').html(res.error.message);
                $('.SignUpBtn').removeClass('processing').attr('disabled', false);
                $('.SignUpBtn').find('span').show();
                $('.SignUpBtn').find('#dvLoading').hide();
                return;
            }
            
            if (res.id) {
                isAjax = true;
                var stripe_customer_id = res.id;
                var data = {
                    action: 'bookbeam_user_signup',
                    plan_type: jQuery('input.plan_input').val(),
                    nonce: book_beam_params.nonce,
                    fullname: jQuery('input.fname').val(),
                    email: jQuery('input.email').val(),
                    password: jQuery('input.password').val(),
                    applied_coupon: jQuery('input.applied_coupon').val(),
                    stripeToken: stripe_customer_id,
                    plan_period: $('.plan_period').val() == 'monthly' || $('.plan_period').val() == 'quarterly' ? $('.plan_period').val() : 'annual',
                }

                $.ajax({

                    url: book_beam_params.ajax_url,
                    type: 'POST',
                    data: data,
                    beforeSend: function () {

                        $('.SignUpBtn').addClass('processing').attr('disabled', true);
                        $('.SignUpBtn').find('span').hide();
                        $('.SignUpBtn').find('#dvLoading').show();

                    },
                    success: function (resp) {

                        if (resp.success && resp.status == 'redirect_to_app') {
                            jQuery('.overlay-container').hide();
                            jQuery('a').unbind("click");
                            window.dataLayer.push({
                                'event' : 'SignUpSubmit',
                                'eventCategory' : 'Sign up',
                                'eventAction' : 'Form Submit'
                              });
                            // jQuery('.sign-up__form .noticeBox').show().addClass('success');
                            // jQuery('.sign-up__form .noticeBox').html(resp.message);

							if($('.SignUpBtn').attr('data-redirect') != ''){
								window.location.href = $('.SignUpBtn').attr('data-redirect');
							}else{
								window.location.href = resp.redirect;
							}
                        }
                        if (resp.success && resp.status == 'pending_for_auth') {
                            stripe.confirmCardPayment(resp.pi_client)
                            .then(result => {
								try{
									if(result.error){
										$('.SignUpBtn').removeClass('processing').attr('disabled', false);
										$('.SignUpBtn').find('span').show();
										$('.SignUpBtn').find('#dvLoading').hide();
										jQuery('.overlay-container').hide();
										jQuery('a').unbind("click");
										jQuery('button').prop('disabled', false);
										return;
									}
								}catch(e){}
                                if(result.paymentIntent.status == 'succeeded'){
                                    //Add ajax for inserting data to both app and site
                                    $.ajax({
                                        url: book_beam_params.ajax_url,
                                        type: 'POST',
                                        data: {
                                            action: 'bookbeam_user_signup_after_3ds',
                                            plan_type: jQuery('input.plan_input').val(),
                                            nonce: book_beam_params.nonce,
                                            fullname: jQuery('input.fname').val(),
                                            email: jQuery('input.email').val(),
                                            password: jQuery('input.password').val(),
                                            applied_coupon: jQuery('input.applied_coupon').val(),
                                            stripeToken: stripe_customer_id,
                                            plan_period: $('.plan_period').val() == 'monthly' ? $('.plan_period').val() : 'annual',
                                            subs_id: resp.subs_id,
                                            customer_id : resp.customer_id
                                        },
                                        success: function (resp) {

                                            if (resp.success) {
                                                jQuery('.overlay-container').hide();
                                                jQuery('a').unbind("click");
                                                window.dataLayer.push({
                                                    'event' : 'SignUpSubmit',
                                                    'eventCategory' : 'Sign up',
                                                    'eventAction' : 'Form Submit'
                                                  });
                                                // jQuery('.sign-up__form .noticeBox').show().addClass('success');
                                                // jQuery('.sign-up__form .noticeBox').html(resp.message);
                    
                                                if($('.SignUpBtn').attr('data-redirect') != ''){
                                                    window.location.href = $('.SignUpBtn').attr('data-redirect');
                                                }else{
                                                    window.location.href = resp.redirect;
                                                }
                                            }

                                            if (!resp.success) {
                                                $('.SignUpBtn').removeClass('processing').attr('disabled', false);
                                                $('.SignUpBtn').find('span').show();
                                                $('.SignUpBtn').find('#dvLoading').hide();
                                                jQuery('.overlay-container').hide();
                                                jQuery('a').unbind("click");
                                                jQuery('button').prop('disabled', false);
                    
                                            }
                                        }
                                    })
                                }else{
                                    $('.SignUpBtn').removeClass('processing').attr('disabled', false);
                                    $('.SignUpBtn').find('span').show();
                                    $('.SignUpBtn').find('#dvLoading').hide();
                                    jQuery('.overlay-container').hide();
                                    jQuery('a').unbind("click");
                                    jQuery('button').prop('disabled', false);
                                }
                            }).catch(err => {
                                console.log({siErr: err});
                            });
                        }
                        if (!resp.success) {
                            $('.SignUpBtn').removeClass('processing').attr('disabled', false);
                            $('.SignUpBtn').find('span').show();
                            $('.SignUpBtn').find('#dvLoading').hide();
                            jQuery('.overlay-container').hide();
                            jQuery('a').unbind("click");
                            jQuery('button').prop('disabled', false);
                            jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                            jQuery('.sign-up__form .noticeBox').html(resp.message);

                        }
                    }
                });
            }
        }
        function bf_token(res){
            
            var $form = jQuery('.sign-up__form');
            // show processing message, disable links and buttons until form is submitted and reloads
            jQuery('a').bind("click", function () { return false; });
            jQuery('button').prop('disabled', true);
            jQuery('.overlay-container').show();

            if (res.error) {
                jQuery('.overlay-container').hide();
                jQuery('a').unbind("click");
                jQuery('button').prop('disabled', false);
                jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                jQuery('.sign-up__form .noticeBox').html(res.error.message);
                $('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
                $('.BFSignUpBtn').find('span').show();
                $('.BFSignUpBtn').find('#dvLoading').hide();
                return;
            }
            
            if (res.id) {
                isAjax = true;
                var stripe_customer_id = res.id;
                var data = {
                    action: 'bf_bookbeam_user_signup',
                    plan_type: jQuery('input.plan_input').val(),
                    nonce: book_beam_params.nonce,
                    fullname: jQuery('input.fname').val(),
                    email: jQuery('input.email').val(),
                    password: jQuery('input.password').val(),
                    stripeToken: stripe_customer_id,
                    plan_period: $('.plan_period').val() == 'monthly' || $('.plan_period').val() == 'quarterly' ? $('.plan_period').val() : 'annual',
                    has_monthly_discount: $('.sign-up__price').hasClass('monthly-coupon-active') ? 1 : 0
                }
                $.ajax({

                    url: book_beam_params.ajax_url,
                    type: 'POST',
                    data: data,
                    beforeSend: function () {

                        $('.BFSignUpBtn').addClass('processing').attr('disabled', true);
                        $('.BFSignUpBtn').find('span').hide();
                        $('.BFSignUpBtn').find('#dvLoading').show();

                    },
                    success: function (resp) {

                        if (resp.success && resp.status == 'redirect_to_app') {
                            jQuery('.overlay-container').hide();
                            jQuery('a').unbind("click");
                            window.dataLayer.push({
                                'event' : 'SignUpSubmit',
                                'eventCategory' : 'Sign up',
                                'eventAction' : 'Form Submit'
                              });
                            // jQuery('.sign-up__form .noticeBox').show().addClass('success');
                            // jQuery('.sign-up__form .noticeBox').html(resp.message);

                            if($('.BFSignUpBtn').attr('data-redirect') != ''){
                                window.location.href = $('.BFSignUpBtn').attr('data-redirect');
                            }else{
                                window.location.href = resp.redirect;
                            }
                        }
                        if (resp.success && resp.status == 'pending_for_auth') {
                            stripe.confirmCardPayment(resp.pi_client)
                            .then(result => {
								try{
									if(result.error){
										$('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
										$('.BFSignUpBtn').find('span').show();
										$('.BFSignUpBtn').find('#dvLoading').hide();
										jQuery('.overlay-container').hide();
										jQuery('a').unbind("click");
										jQuery('button').prop('disabled', false);
										return;
									}
								}catch(e){}
                                if(result.paymentIntent.status == 'succeeded'){
                                    //Add ajax for inserting data to both app and site
                                    $.ajax({
                                        url: book_beam_params.ajax_url,
                                        type: 'POST',
                                        data: {
                                            action: 'bf_bookbeam_user_signup_after_3ds',
                                            plan_type: jQuery('input.plan_input').val(),
                                            nonce: book_beam_params.nonce,
                                            fullname: jQuery('input.fname').val(),
                                            email: jQuery('input.email').val(),
                                            password: jQuery('input.password').val(),
                                            stripeToken: stripe_customer_id,
                                            plan_period: $('.plan_period').val() == 'monthly' ? $('.plan_period').val() : 'annual',
                                            subs_id: resp.subs_id,
                                            customer_id : resp.customer_id
                                        },
                                        success: function (resp) {

                                            if (resp.success) {
                                                jQuery('.overlay-container').hide();
                                                jQuery('a').unbind("click");
                                                window.dataLayer.push({
                                                    'event' : 'SignUpSubmit',
                                                    'eventCategory' : 'Sign up',
                                                    'eventAction' : 'Form Submit'
                                                  });
                                                // jQuery('.sign-up__form .noticeBox').show().addClass('success');
                                                // jQuery('.sign-up__form .noticeBox').html(resp.message);
                    
                                                if($('.BFSignUpBtn').attr('data-redirect') != ''){
                                                    window.location.href = $('.BFSignUpBtn').attr('data-redirect');
                                                }else{
                                                    window.location.href = resp.redirect;
                                                }
                                            }

                                            if (!resp.success) {
                                                $('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
                                                $('.BFSignUpBtn').find('span').show();
                                                $('.BFSignUpBtn').find('#dvLoading').hide();
                                                jQuery('.overlay-container').hide();
                                                jQuery('a').unbind("click");
                                                jQuery('button').prop('disabled', false);
                    
                                            }
                                        }
                                    })
                                }else{
                                    $('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
                                    $('.BFSignUpBtn').find('span').show();
                                    $('.BFSignUpBtn').find('#dvLoading').hide();
                                    jQuery('.overlay-container').hide();
                                    jQuery('a').unbind("click");
                                    jQuery('button').prop('disabled', false);
                                }
                            }).catch(err => {
                                console.log({siErr: err});
                            });
                        }
                        if (!resp.success) {
                            $('.BFSignUpBtn').removeClass('processing').attr('disabled', false);
                            $('.BFSignUpBtn').find('span').show();
                            $('.BFSignUpBtn').find('#dvLoading').hide();
                            jQuery('.overlay-container').hide();
                            jQuery('a').unbind("click");
                            jQuery('button').prop('disabled', false);
                            jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                            jQuery('.sign-up__form .noticeBox').html(resp.message);

                        }
                    }
                });
            }
        }
    } catch (error) {}
    // Remove warning class when user types in the input box
    $('.sign-up__form input.textBox').on('blur focusout', function () {
        if ($(this).val() != '') {
            $(this).removeClass('warning');
            $(this).parent().removeClass('warning');
            $(this).parent().removeClass('error');
            $(this).parent().find('.input-error').html('').hide();
            $('.sign-up__form .noticeBox').hide();
        }
    });
    $('body').on('click', '.freeSignUpBtn', function(e){
        e.preventDefault();
        var $this = $(this);
        $('.sign-up__form input.textBox').each(function () {

            if ($(this).val() === '') {
                if ($(this).hasClass('required')) {
                    $(this).addClass('warning');
                    $(this).parent().addClass('warning');

                    $(this).parent().find('.input-error').show();
                    $(this).parent().find('.input-error').html('This field is required should not be empty.');
                }
            }

        });

        if (!$('.sign-up__agreement_checkbox').is(':checked')) {
            $('.sign-up__agreement_checkbox').parent().find('.input-error').show()
            $('.sign-up__agreement_checkbox').parent().find('.input-error').html('Please tick the box before proceeding.');
            $('.sign-up__agreement_checkbox').parent().addClass('warning')
        } else {
            $('.sign-up__agreement_checkbox').parent().find('.input-error').hide()
            $('.sign-up__agreement_checkbox').parent().removeClass('warning')
        }
        if (checkemailvalidation($('.email').val()) === false) {

            $('.email').parent().find('.input-error').show();
            $('.email').parent().find('.input-error').html('Please enter a valid email address');
            $('.email').parent().addClass('warning');
            $('.email').addClass('warning');

        }

        if ($('.password').val() !== $('.confirmPassword').val()) {

            $('.confirmPassword').parent().find('.input-error').show();
            $('.confirmPassword').parent().find('.input-error').html('Password does not match');
            $('.confirmPassword').addClass('warning');
            $('.confirmPassword').parent().addClass('warning');
            $('.password').addClass('warning');

        }


        if ($('.fname').val() !== '' && checkemailvalidation($('.email').val()) === true) {
            if ($('.password').val() !== '' && $('.confirmPassword').val() !== '') {
                if ($('.password').val() === $('.confirmPassword').val() && $('.sign-up__agreement_checkbox').is(':checked')) {
                    
                    if (!$this.hasClass('processing')) {
                    
                        $this.addClass('processing');
                        $('.sign-up__form .noticeBox').hide();
                        $('.sign-up__form input.textBox').removeClass('warning');
                        $('.sign-up__form label .input-error').hide().html('');
                        var data = {
                            action: 'bookbeam_free_signup',
                            nonce: book_beam_params.nonce,
                            fullname: jQuery('input.fname').val(),
                            email: jQuery('input.email').val(),
                            password: jQuery('input.password').val(),
                        }
                        $.ajax({
                            url: book_beam_params.ajax_url,
                            type: 'POST',
                            data: data,
                            beforeSend: function(){
                                $this.addClass('processing').attr('disabled', true);
                                $this.find('span').hide();
                                $this.find('#dvLoading').show();
                            },
                            success: function(response){
                                if(!response.success){        
                                    $('.freeSignUpBtn').removeClass('processing').attr('disabled', false);
                                    $('.freeSignUpBtn').find('span').show();
                                    $('.freeSignUpBtn').find('#dvLoading').hide();
                                    jQuery('.overlay-container').hide();
                                    jQuery('a').unbind("click");
                                    jQuery('button').prop('disabled', false);
                                    jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                                    jQuery('.sign-up__form .noticeBox').html(response.message);
                                    $this.find('span').show();
                                    $this.find('#dvLoading').hide();
                                }else{
                                    window.location.href = site_url + '/thank-you';
                                }
                            }

                        })
                    }
                }
            }
        }
    })
	
    $('body').on('click', '.freeSignUpBtn2', function(e){
        e.preventDefault();
        var $this = $(this);
        $('.sign-up__form input.textBox').each(function () {

            if ($(this).val() === '') {
                if ($(this).hasClass('required')) {
                    $(this).addClass('warning');
                    $(this).parent().addClass('warning');

                    $(this).parent().find('.input-error').show();
                    $(this).parent().find('.input-error').html('This field is required should not be empty.');
                }
            }

        });

        if (!$('.sign-up__agreement_checkbox').is(':checked')) {
            $('.sign-up__agreement_checkbox').parent().find('.input-error').show()
            $('.sign-up__agreement_checkbox').parent().find('.input-error').html('Please tick the box before proceeding.');
            $('.sign-up__agreement_checkbox').parent().addClass('warning')
        } else {
            $('.sign-up__agreement_checkbox').parent().find('.input-error').hide()
            $('.sign-up__agreement_checkbox').parent().removeClass('warning')
        }
        if (checkemailvalidation($('.email').val()) === false) {

            $('.email').parent().find('.input-error').show();
            $('.email').parent().find('.input-error').html('Please enter a valid email address');
            $('.email').parent().addClass('warning');
            $('.email').addClass('warning');

        }

        if ($('.password').val() !== $('.confirmPassword').val()) {

            $('.confirmPassword').parent().find('.input-error').show();
            $('.confirmPassword').parent().find('.input-error').html('Password does not match');
            $('.confirmPassword').addClass('warning');
            $('.confirmPassword').parent().addClass('warning');
            $('.password').addClass('warning');

        }


        if ($('.fname').val() !== '' && checkemailvalidation($('.email').val()) === true) {
            if ($('.password').val() !== '' && $('.confirmPassword').val() !== '') {
                if ($('.password').val() === $('.confirmPassword').val() && $('.sign-up__agreement_checkbox').is(':checked')) {
                    
                    if (!$this.hasClass('processing')) {
                    
                        $this.addClass('processing');
                        $('.sign-up__form .noticeBox').hide();
                        $('.sign-up__form input.textBox').removeClass('warning');
                        $('.sign-up__form label .input-error').hide().html('');
                        var data = {
                            action: 'bookbeam_free_signup',
                            nonce: book_beam_params.nonce,
                            fullname: jQuery('input.fname').val(),
                            email: jQuery('input.email').val(),
                            password: jQuery('input.password').val(),
                            fromPopup: 1,
                        }
                        $.ajax({
                            url: book_beam_params.ajax_url,
                            type: 'POST',
                            data: data,
                            beforeSend: function(){
                                $this.addClass('processing').attr('disabled', true);
                                $this.find('span').hide();
                                $this.find('#dvLoading').show();
                            },
                            success: function(response){
                                if(!response.success){        
                                    $('.freeSignUpBtn').removeClass('processing').attr('disabled', false);
                                    $('.freeSignUpBtn').find('span').show();
                                    $('.freeSignUpBtn').find('#dvLoading').hide();
                                    jQuery('.overlay-container').hide();
                                    jQuery('a').unbind("click");
                                    jQuery('button').prop('disabled', false);
                                    jQuery('.sign-up__form .noticeBox').show().removeClass('success');
                                    jQuery('.sign-up__form .noticeBox').html(response.message);
                                    $this.find('span').show();
                                    $this.find('#dvLoading').hide();
                                }else{
                                    window.location.href = site_url + '/thank-you';
                                }
                            }

                        })
                    }
                }
            }
        }
    })
    $('body').on('click', 'a[href*=chrome-extension-lite-sign-up]', function(e){
        e.preventDefault();
        e.stopPropagation();
        window.location.href = site_url + '/chrome-extension-lite-sign-up';
    })
    $('body').on('click', 'a[href*=ce-lite-signup]', function(e){
        e.preventDefault();
        e.stopPropagation();
        window.location.href = site_url + '/ce-lite-signup';
    })
    $('body').on('click', '.plans__info:not(.from-thankyou) .button', function (e) {
        e.preventDefault();

        var $this = $(this);

        $this.find('span').hide();
        $this.find('.sk-circle').show();
        //Check plan_type
        var pricing = $('.pricing__switch').length > 0 ? $('.pricing__switch').attr('data-pay-switch') || 'annual' : $('.bf_pricing__switch').attr('data-pay-switch') || 'annual';
        var plan_title = $(this).parent().find('h4').text();
        var monthly_price = $(this).parent().find('.plans__price .monthly').text();
        var yearly_price = $(this).parent().find('.plans__price .yearly').text();
        var old_price = $(this).parent().find('.plans__old-price').text();
        var quarterly_price = $(this).parent().find('.plans__price .quarterly') ? $(this).parent().find('.plans__price .quarterly').text() : '';
        var original_monthly_d_price = yearly_price;
        var original_yearly_price = 0;
        var yearly_del_text = $(this).parent().find('.plans__price .yearly del').text();
        var quarterly_del_text = $(this).parent().parent().find('.plans__price .quarterly') ? $(this).parent().parent().find('.plans__price .quarterly del').text() : '';
        var quarterly_saved = $($(this).parent().find('.plans__saved-price.quarterly div')[$(this).parent().find('.plans__saved-price.quarterly div').length - 1]).text();
        var monthly_saved = $($(this).parent().find('.plans__saved-price.monthly div')[$(this).parent().find('.plans__saved-price.monthly div').length - 1]).text();
        var yearly_saved = $($(this).parent().find('.plans__saved-price.annually div')[$(this).parent().find('.plans__saved-price.annually div').length - 1]).text();
        if (pricing == 'yearly') {
            pricing = 'annual';
        }
        //remove /mo
        monthly_price = monthly_price.replace('/mo', '');
        yearly_price = yearly_price.replace('/mo', '');
        old_price = old_price.replace('/mo', '');

        yearly_price = yearly_price.replace(/\$/g, '');
        yearly_del_text = yearly_del_text.replace('$', '');
        monthly_price = monthly_price.replace('$', '');
        old_price = old_price.replace('$', '');
        
        if($this.parent().hasClass('blackfriday')){
            yearly_price = yearly_price.replace('for the first year', '');
            yearly_price = yearly_price.replace(yearly_del_text, '');
            monthly_price = monthly_price.replace('for the first 6 months', '');
            yearly_price = (parseFloat(yearly_price) / 12).toFixed(2);
            original_monthly_d_price = '$' + yearly_price + '/mo';
            var prices_yearly = $('.plans__price .yearly');
            var prices_monthly = $('.plans__price .monthly');

            if(quarterly_price != ''){
                quarterly_price = quarterly_price.replace(quarterly_del_text, '');
            }

            if (data_prices_yearly.length === 0) {
                prices_yearly.each(function (i, e) {
                    data_prices_yearly.push($(e).text());
                })
            }
            if (data_prices_monthly.length === 0) {
                prices_monthly.each(function (i, e) {
                    var price = $(e).text();
                    price = price.replace('/mo', '');
                    price = price.replace('$', '');
                    // price = parseFloat(price) * 2; //Uncomment if monthly has discount for black friday
                    price = parseFloat(price) * 1; //Remove if monthly has discount for black friday
                    price = '$' + price + '/mo';
                    data_prices_monthly.push(price);
                })
            }

        }
        //Trim remove white space
        monthly_price = +monthly_price.trim();
        yearly_price = +yearly_price.trim();
        old_price = old_price.trim();
        var plan_type = '';
        plan_title = plan_title.replace('New', '').trim();
        if (plan_title == 'Starter') {
            plan_type = 'Starter';
        }
        if (plan_title == 'Basic') {
            plan_type = 'basic';
        }
        if (plan_title == 'Pro') {
            plan_type = 'pro';
        }
        if(plan_title == 'Publisher Pro') {
            plan_type = 'publisherpro';
        }
        if(plan_title == 'Publisher'){
            plan_type = 'publisher';
        }
        // window.location.href = site_url + '/registration/?plan_type=' + plan_type + '&pricing=' + pricing + '&plan_title=' + plan_title + '&monthly_price=' + monthly_price + '&yearly_price=' + yearly_price + '&old_price=' + old_price;

        if (pricing == 'annual') {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[$(this).parent().parent().index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + Number(original_yearly_price.toFixed(2));
            old_price = '$' + old_price + '/mo';
            monthly_price = '$' + monthly_price;

        } else {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[$(this).parent().parent().index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + Number(original_yearly_price.toFixed(2));
            old_price = '$' + old_price + '/mo';

            monthly_price = '$' + monthly_price;
        }

        yearly_price = yearly_price * 12;
        yearly_price = parseFloat(yearly_price.toFixed(2));
        if(yearly_price % 1 !== 0){
            yearly_price = '$' + yearly_price.toFixed(2);
        }else{
            yearly_price = '$' + yearly_price;
        }
        if (!$this.hasClass('processing')) {
            // $this.addClass('processing');
            $('.plans__info .button').addClass('processing');
            var isQuarterly = 0;
            
            if($('main').hasClass('pricing-v2') && plan_title != 'Starter' && pricing != 'monthly'){
                isQuarterly = 1;
            }
            var data = {
                    plan_type: plan_type,
                    pricing: pricing,
                    plan_title: plan_title,
                    monthly_price: monthly_price,
                    original_yearly_price: original_yearly_price,
                    original_monthly_d_price: original_monthly_d_price,
                    original_quarterly_price: quarterly_del_text,
                    yearly_price: yearly_price,
                    old_price: old_price,
                    is_quarterly: isQuarterly,
                    quarterly_price: quarterly_price,
                    quarterly_saved: quarterly_saved,
                    monthly_saved: monthly_saved,
                    yearly_saved: yearly_saved,
                    coupon: $('.coupon_active').val() != '' ? $('.coupon_active').val() : '',
                    coupon_type: $('.coupon_type').val() != '' ? $('.coupon_type').val() : '',
            }
            if($this.hasClass('bf-button')){
                Cookies.set('bookbeam_pricing_data_bf', JSON.stringify(data), { expires: 1, path: '/' });
            }else{
                Cookies.set('bookbeam_pricing_data', JSON.stringify(data), { expires: 30, path: '/' });
            }
            window.dataLayer.push({
                'event' : 'PricingToRegistrationClick',
                'eventCategory' : 'Button',
                'eventAction' : 'Click'
              });
            
            
            $this.find('span').show();
            $this.find('.sk-circle').hide();
            $('.plans__info .button').removeClass('processing');

            // if($('body').hasClass('early-bird-special')){
            //     window.location.href = site_url + '/early-bird-special-registration/';
            //     return;
            // }
            // if (window.location.pathname.includes('black-friday')) {
            //     // window.location.href = site_url + '/registration';
            //     window.location.href = site_url + '/black-friday-registration/';
            //     return;
            // }else{
            //     window.location.href = site_url + '/registration';
            //     return;
            //     // window.location.href = site_url + '/registration-wp-stripe-form';
            // }
            let bbParameter = '?plan=' + plan_title.replace(' ', '') + ' ' + pricing;
            if(Cookies.get('affwp_ref')){
                bbParameter += '&affiliate=' + Cookies.get('affwp_ref');
            }
            if($('.coupon_active').val() != ""){
                bbParameter += '&coupon=' + $('.coupon_active').val();
            }
            let url = "https://app.zB2mCiISUVux.io/#/auth/signup" //https://qa.zB2mCiISUVux.io/#/auth/signup
            window.location.href = url + bbParameter;
            
        }

    })
    $('body').on('click', '.plans__features:not(.from-thankyou) .button', function (e) {
        e.preventDefault();

        var $this = $(this);

        $this.find('span').hide();
        $this.find('.sk-circle').show();
        
        //Check plan_type
        var pricing = $('.pricing__switch').length > 0 ? $('.pricing__switch').attr('data-pay-switch') || 'annual' : $('.bf_pricing__switch').attr('data-pay-switch') || 'annual';
        var plan_title = $(this).parent().parent().find('h4').text();
        var monthly_price = $(this).parent().parent().find('.plans__price .monthly').text();
        var yearly_price = $(this).parent().parent().find('.plans__price .yearly').text();
        var old_price = $(this).parent().parent().find('.plans__old-price').text();
        var quarterly_price = $(this).parent().parent().find('.plans__price .quarterly') ? $(this).parent().parent().find('.plans__price .quarterly').text() : '';
        var original_monthly_d_price = yearly_price;
        var original_yearly_price = 0;
        var yearly_del_text = $(this).parent().parent().find('.plans__price .yearly del').text();
        var quarterly_del_text = $(this).parent().parent().find('.plans__price .quarterly') ? $(this).parent().parent().find('.plans__price .quarterly del').text() : '';
        var quarterly_saved = $($(this).parent().parent().find('.plans__saved-price.quarterly div')[$(this).parent().find('.plans__saved-price.quarterly div').length - 1]).text();
        var monthly_saved = $($(this).parent().parent().find('.plans__saved-price.monthly div')[$(this).parent().find('.plans__saved-price.monthly div').length - 1]).text();
        var yearly_saved = $($(this).parent().parent().find('.plans__saved-price.annually div')[$(this).parent().find('.plans__saved-price.annually div').length - 1]).text();
        
        if (pricing == 'yearly') {
            pricing = 'annual';
        }
        //remove /mo
        monthly_price = monthly_price.replace('/mo', '');
        yearly_price = yearly_price.replace('/mo', '');
        old_price = old_price.replace('/mo', '');

        yearly_price = yearly_price.replace(/\$/g, '');
        yearly_del_text = yearly_del_text.replace('$', '');
        monthly_price = monthly_price.replace('$', '');
        old_price = old_price.replace('$', '');
        
        if($this.parent().parent().find('.plans__info').hasClass('blackfriday')){
            yearly_price = yearly_price.replace('for the first year', '');
            
            yearly_price = yearly_price.replace(yearly_del_text, '');
            monthly_price = monthly_price.replace('for the first 6 months', '');
            yearly_price = (parseFloat(yearly_price) / 12).toFixed(2);
            original_monthly_d_price = '$' + yearly_price + '/mo';

            if(quarterly_price != ''){
                quarterly_price = quarterly_price.replace(quarterly_del_text, '');
            }

            var prices_yearly = $('.plans__price .yearly');
            var prices_monthly = $('.plans__price .monthly');
            if (data_prices_yearly.length === 0) {
                prices_yearly.each(function (i, e) {
                    data_prices_yearly.push($(e).text());
                })
            }
            if (data_prices_monthly.length === 0) {
                prices_monthly.each(function (i, e) {
                    var price = $(e).text();
                    price = price.replace('/mo', '');
                    price = price.replace('$', '');
                    // price = parseFloat(price) * 2; //Uncomment if monthly has discount for black friday
                    price = parseFloat(price) * 1; //Remove if monthly has discount for black friday
                    price = '$' + price + '/mo';
                    data_prices_monthly.push(price);
                })
            }

        }
        //Trim remove white space
        monthly_price = +monthly_price.trim();
        yearly_price = +yearly_price.trim();
        old_price = old_price.trim();
        var plan_type = '';
        if (plan_title == 'Starter') {
            plan_type = 'Starter';
        }
        if (plan_title == 'Basic') {
            plan_type = 'basic';
        }
        if (plan_title == 'Pro') {
            plan_type = 'pro';
        }
        if(plan_title == 'Publisher Pro'){
            plan_type = 'publisherpro'
        }
        if(plan_title == 'Publisher'){
            plan_type = 'publisher';
        }

        // window.location.href = site_url + '/registration/?plan_type=' + plan_type + '&pricing=' + pricing + '&plan_title=' + plan_title + '&monthly_price=' + monthly_price + '&yearly_price=' + yearly_price + '&old_price=' + old_price;

        if (pricing == 'annual') {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[$(this).parent().parent().index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + Number(original_yearly_price.toFixed(2));
            old_price = '$' + old_price + '/mo';
            monthly_price = '$' + monthly_price;

        } else {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[$(this).parent().parent().index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + Number(original_yearly_price.toFixed(2));
            old_price = '$' + old_price + '/mo';

            monthly_price = '$' + monthly_price;
        }

        yearly_price = yearly_price * 12;
        yearly_price = parseFloat(yearly_price.toFixed(2));
        if(yearly_price % 1 !== 0){
            yearly_price = '$' + yearly_price.toFixed(2);
        }else{
            yearly_price = '$' + yearly_price;
        }
        if (!$this.hasClass('processing')) {
            // $this.addClass('processing');
            $('.plans__info .button').addClass('processing');
            var isQuarterly = 0;
            
            if($('main').hasClass('pricing-v2') && plan_title != 'Starter' && pricing != 'monthly'){
                isQuarterly = 1;
            }
            var data = {
                    plan_type: plan_type,
                    pricing: pricing,
                    plan_title: plan_title,
                    monthly_price: monthly_price,
                    original_yearly_price: original_yearly_price,
                    original_monthly_d_price: original_monthly_d_price,
                    original_quarterly_price: quarterly_del_text,
                    yearly_price: yearly_price,
                    old_price: old_price,
                    is_quarterly: isQuarterly,
                    quarterly_price: quarterly_price,
                    quarterly_saved: quarterly_saved,
                    monthly_saved: monthly_saved,
                    yearly_saved: yearly_saved,
                    coupon: $('.coupon_active').val() != '' ? $('.coupon_active').val() : '',
                    coupon_type: $('.coupon_type').val() != '' ? $('.coupon_type').val() : '',
            }
            if($this.hasClass('bf-button')){
                Cookies.set('bookbeam_pricing_data_bf', JSON.stringify(data), { expires: 1, path: '/' });
            }else{
                Cookies.set('bookbeam_pricing_data', JSON.stringify(data), { expires: 30, path: '/' });
            }
            window.dataLayer.push({
                'event' : 'PricingToRegistrationClick',
                'eventCategory' : 'Button',
                'eventAction' : 'Click'
              });

            
            $this.find('span').show();
            $this.find('.sk-circle').hide();
            $('.plans__info .button').removeClass('processing');

            $('.plans__info .button').addClass('processing');
            // if($('body').hasClass('early-bird-special')){
            //     window.location.href = site_url + '/early-bird-special-registration/';
            //     return;
            // }
            // if (window.location.pathname.includes('black-friday')) {
            //     // window.location.href = site_url + '/registration';
            //     window.location.href = site_url + '/black-friday-registration/';
            //     return;
            // }else{
            //     window.location.href = site_url + '/registration';
            //     return;
            //     // window.location.href = site_url + '/registration-wp-stripe-form';
            // }

            let bbParameter = '?plan=' + plan_title.replace(' ', '') + ' ' + pricing;
            if(Cookies.get('affwp_ref')){
                bbParameter += '&affiliate=' + Cookies.get('affwp_ref');
            }
            if($('.coupon_active').val() != ""){
                bbParameter += '&coupon=' + $('.coupon_active').val();
            }
            let url = "https://app.zB2mCiISUVux.io/#/auth/signup" //https://qa.zB2mCiISUVux.io/#/auth/signup
            window.location.href = url + bbParameter;
        }

    })

    let urlParams = new URLSearchParams(window.location.search);
    let urlSiteCoupon = urlParams.get('couponCode');
    var cookieValue = Cookies.get('siteCouponCookie');
    if(urlSiteCoupon){
        if(!cookieValue && cookieValue != urlSiteCoupon){
            var data = {
                'action': 'checkCoupon',
                'coupon': urlSiteCoupon,
                'coupon_type': 'site_coupon'
            };

            $.ajax({
                url: book_beam_params.ajax_url,
                type: 'POST',
                data: data,
                beforeSend: function (e) {
                },
                success: function (response) {
                    if(response.affiliate > 0){ 
                        let ribbon = '';
                        urlSiteCoupon = urlSiteCoupon.toUpperCase();
                        ribbon = '<div id="coupon_ribbon" style="text-align: center; display: block;"><p><strong>Coupon <b style="text-transform:uppercase;">"' + urlSiteCoupon + '"</b> applied.  Save <u>up to 44% with annual</u>, up to 22% for quarterly or 15% for monthly for 6 months!</strong></p></div>';
                        if (ribbon !== '') {
                            $('.announcement-ribbon').html(ribbon);
                        }
                        Cookies.set('siteCouponCookie', urlSiteCoupon, { expires: 1 });//On Live
                        // Cookies.set('siteCouponCookie', urlSiteCoupon, { expires: 5 / (24 * 60) });
                    }
                }
            });
        }
    }
    if(cookieValue){
        let ribbon = '';
        cookieValue = cookieValue.toUpperCase();
        ribbon = '<div id="coupon_ribbon" style="text-align: center; display: block;"><p><strong>Coupon <b style="text-transform:uppercase;">"' + cookieValue + '"</b> applied.  Save <u>up to 44% with annual</u>, up to 22% for quarterly or 15% for monthly for 6 months!</strong></p></div>';
        if (ribbon !== '') {
            $('.announcement-ribbon').html(ribbon);
        }
    }


    $('span.coupon-price').hide();
    $('body').on('click', '.pricing__switch li', function (e) {
      var $coupon = jQuery('input.coupon__input').val() || $('.coupon_active').val();
      if ($('.coupon_type').val() != '') {
          $('.plans__old-price').css('visibility', 'visible');
          var pay_switch = $('.pricing__switch').attr('data-pay-switch');
          var quarterly_msg = ''
          if($('.plans__cta.quarterly').length > 0){
              quarterly_msg = ', up to 22% for quarterly';
          }
          if ($('.coupon_type').val() == 50) {
              if (pay_switch == 'monthly') {
                  var msg = 'Coupon <b style="text-transform:uppercase;">' + $coupon + '</b> applied. Save up to 44% with annual'+ quarterly_msg+' or 15% for monthly for 6 months!';
              } else {
                  var msg = 'Coupon <b style="text-transform:uppercase;">' + $coupon + '</b> applied. Save up to 44% with annual'+ quarterly_msg+' or 15% for monthly for 6 months!';
              }
              $('.noticeBox').removeClass('no-discount');
          } else {
              if (pay_switch == 'monthly') {
                  $('.plans__old-price').css('visibility', 'hidden');
                  $('.noticeBox').addClass('no-discount');
                  var msg = 'Coupon <b style="text-transform:uppercase;">' + $coupon + '</b> applied. Save up to 44% with annual'+ quarterly_msg+' or 15% for monthly for 6 months!';
              } else {
                  var msg = 'Coupon <b style="text-transform:uppercase;">' + $coupon + '</b> applied. Save up to 44% with annual'+ quarterly_msg+' or 15% for monthly for 6 months!';
                  $('.noticeBox').removeClass('no-discount');
              }
          }
          $('.noticeBox').html(msg).show();
          if(pay_switch != 'quarterly'){
              $('.plans__cta:not(.quarterly)').removeClass('hidden')
          }else{
              $('._plans__cta:not(.quarterly)').addClass('hidden');
          }
      }

      var pay_switch = $('.pricing__switch').attr('data-pay-switch');
      if (pay_switch == 'monthly') {
          if($('.coupon_type').val() != ''){
              $('.plans__cta:not(.quarterly)').each(function(e){
                  if($(this).parent().find('.plans__stored_cta').val() == ''){
                      $(this).parent().find('.plans__stored_cta').val($(this).text());
                  }
                  $(this).text('15% Off for 6 Months');
                  $(this).removeClass('hidden')
              })
          }
      } else if (pay_switch == 'yearly') {
          // $('.plans__cta').removeClass('hidden');
          if($('.coupon_type').val() != ''){
              $('.plans__cta:not(.quarterly)').each(function(e){
                  if($(this).parent().find('.plans__stored_cta').val() != ''){
                      var cta_value = $(this).parent().find('.plans__stored_cta').val();
                      $(this).text(cta_value);
                      if($('.coupon_type').val() == 50){
                          $(this).removeClass('hidden')
                      }
                  }
              })
          }
      }
    })
    $('body').on('click', '.bf_pricing__switch li', function (e) {
        var pay_switch = $('.bf_pricing__switch').attr('data-pay-switch');
        if (pay_switch == 'monthly') {
            $('.plans__cta').addClass('hidden');
        } else if (pay_switch == 'yearly') {
            $('.plans__cta').removeClass('hidden');
        }
    })
    $('.sign-up__form .affiliateCoupon').on('keypress input paste propertychange', function (e) {
        // if ($('.applied_coupon').val() != '') {
        //     $('.applied_coupon').val($(this).val());
        // }

        if ($(this).val() != '') {
            $('.sign-up__form .apply-coupon-container .button-container').show();
        } else {
            $('.sign-up__form .apply-coupon-container .button-container').hide();
        }
    })
    $('.sign-up__form .apply-coupon-container .button-container a').on('click', function (e) {
        e.preventDefault();
        var $this = $(this);
        var coupon = $('.sign-up__form .affiliateCoupon').val();
        var plan_period = $('.sign-up__form .plan_period').val();
        var old_price = $('.sign-up__form .old_price_p_month').val();
        var old_yearly_price_p_month = $('.sign-up__form .old_yearly_price_p_month').val();
        var plan_type = $('.sign-up__form .plan_input').val();
        if (!$this.hasClass('processing')) {
            $this.addClass('processing');

            var data = {
                'action': 'checkCoupon',
                'coupon': coupon,
            };

            $.ajax({
                url: book_beam_params.ajax_url,
                type: 'POST',
                data: data,
                beforeSend: function (e) {
                    $this.hide();
                    $this.parent().find('.spin-loader').show();
                    $('.sign-up__price').hide();
                    $('.sign-up__cta').hide();
                    $('.sign-up__wrap .spin-loader').show();
                },
                success: function (response) {

                    if (response.affiliate > 0) {
                        var couponType = response.couponType;
                        old_price = old_price.replace('$', '');
                        old_price = old_price.replace('/mo', '');
                        old_price = +old_price.trim();

                        old_yearly_price_p_month = old_yearly_price_p_month.replace('$', '');
                        old_yearly_price_p_month = old_yearly_price_p_month.replace('/mo', '');
                        old_yearly_price_p_month = +old_yearly_price_p_month.trim();

                        //get old yearly_price per month
                        yearly_to_monthly = old_yearly_price_p_month / 12;
                        var signup_el = $('section.sign-up');
                        if(signup_el.hasClass('is-quarterly')){
                            var q_saved_discount = 20;
                            var saved_quarterly = $('.sign-up__price .inactive.quarterly').text();
                            saved_quarterly = saved_quarterly.replace('(Save $', '');
                            saved_quarterly = parseFloat(saved_quarterly.replace(')', ''));
                            if(signup_el.attr('data-plan') == 'publisher-pro'){
                                q_saved_discount = 38;
                            }
                            var discounted_quarterly = q_saved_discount + saved_quarterly;
                            $('.sign-up__price .inactive.quarterly').text('(Save $' + discounted_quarterly + ')');
                        }
                        if ($('.plan_period').val() == 'monthly') {
                            if (couponType == '50') {
                                var discounted_price = old_price - (old_price * .15);
                            } else {
                                var discounted_price = old_price;
                            }

                        } else {
                            var discounted_price = yearly_to_monthly - (yearly_to_monthly * .10)
                        }

                        if ($('.plan_period').val() == 'monthly') {
                            $('.sign-up__cta .cta__monthly ').addClass('active');
                            if (couponType == '50') {
                                if ($('.sign-up__price .annual-orig-price').length > 0 && !$('.sign-up__price').hasClass('has-coupon')) {
                                    var discounted_annual = yearly_to_monthly - (yearly_to_monthly * .10);
                                    annual_price = Number(discounted_annual.toFixed(2)) * 12;
                                    annual_price = annual_price.toFixed(2);

                                    $('.sign-up__price .inactive.annual').text('($' + discounted_annual.toFixed(2) + '/mo)');
                                    $('.sign-up__price .active .annual').text('$' + annual_price);

                                    $('.sign-up__price .annual-orig-price').text('$' + (old_price * 12));
                                } else {
                                    if ($('.sign-up__price .annual-orig-price').length === 0) {

                                        $('<div class="annual-orig-price">$' + old_price + '/mo</div>').insertBefore('.sign-up__price .active');
                                    }
                                }
                                // $('.sign-up__price').addClass('has-coupon')
                                // $('.monthly-orig-price').show();
                                if($this.hasClass('stripe-custom-coupon')){
                                    $('.stripe__payment-form input[name="simpay_field[coupon]"]').val('monthly-coupon');
                                    $('.stripe__payment-form .simpay-apply-coupon').click();
                                }
                            } else {
                                if ($('.sign-up__price .annual-orig-price').length > 0 && !$('.sign-up__price').hasClass('has-coupon')) {
                                    var discounted_annual = yearly_to_monthly - (yearly_to_monthly * .10);
                                    annual_price = Number(discounted_annual.toFixed(2)) * 12;
                                    annual_price = annual_price.toFixed(2);

                                    $('.sign-up__price .inactive.annual').text('($' + discounted_annual.toFixed(2) + '/mo)');
                                    $('.sign-up__price .active .annual').text('$' + annual_price);

                                    $('.sign-up__price .annual-orig-price').text('$' + (old_price * 12));

                                    var discounted_price = old_price - (old_price * .15);
                                    
                                    // $('.sign-up__price .active .monthly').text('$' + monthly_price.toFixed(2));
                                }
                                // $('.monthly-orig-price').hide();
                                if($this.hasClass('stripe-custom-coupon')){
                                    $('.stripe__payment-form input[name="simpay_field[coupon]"]').val('annual-coupon');
                                    $('.stripe__payment-form .simpay-apply-coupon').click();
                                }
                            }
                            $('.sign-up__price').addClass('has-coupon')

                            $('.sign-up__price .active .monthly').text('$' + discounted_price);
                        } else {
                            $('.sign-up__cta .cta__yearly').addClass('active');
                            if($('.sign-up').hasClass('is-quarterly')){
                                $('.sign-up__cta .cta__quarterly').removeClass('active');
                            }
                        
                            if (!$('.sign-up__price').hasClass('has-coupon')) {
                                var discounted_price_year = Number(discounted_price.toFixed(2)) * 12;
                                discounted_price_year = discounted_price_year.toFixed(2);

                                $('.sign-up__price .inactive.annual').text('($' + discounted_price.toFixed(2) + '/mo)');
                                $('.sign-up__price .active .annual').text('$' + discounted_price_year);
                                
                                var monthly_price = old_price - (old_price * .15);
                                
                                $('.sign-up__price .active .monthly').text('$' + monthly_price.toFixed(2));
                                $('.sign-up__price').addClass('has-coupon')
                            }
                            // $('.monthly-orig-price').show();
                            if($this.hasClass('stripe-custom-coupon')){
                                $('.stripe__payment-form input[name="simpay_field[coupon]"]').val('annual-coupon');
                                $('.stripe__payment-form .simpay-apply-coupon').click();
                            }
                        }
                        if($('.sign-up').hasClass('is-quarterly')){
                            var quarterly_discount = 20;
                            if(plan_type === 'publisher-pro'){
                                quarterly_discount = 38;
                            }
                            // var montlhy_price_orig = $('.sign-up__price .monthly-orig-price').text();
                            // monthly_price_orig = montlhy_price_orig.replace('/mo', '');
                            // monthly_price_orig = monthly_price_orig.trim();
                            // monthly_price_orig = monthly_price_orig.replace('$', '');
                            // monthly_price_orig = +monthly_price_orig.trim();
                            // var quarterly_orig_price = (monthly_price_orig * 3) - 10;
                            var quarterly_orig_price = $('.sign-up__price .quarterly-price').text();
                            quarterly_orig_price = quarterly_orig_price.replace('for 3 months', '');
                            quarterly_orig_price = quarterly_orig_price.trim();
                            quarterly_orig_price = quarterly_orig_price.replace('$', '');
                            quarterly_orig_price = +quarterly_orig_price.trim();
                            quarterly_orig_price = quarterly_orig_price - quarterly_discount;
                            $('.sign-up__price .active .quarterly-price').html('$'+ quarterly_orig_price +' <span class="smaller">for 3 months</span>');
                            
                            if($('.plan_period').val() == 'quarterly'){
                                $('.sign-up__cta .cta__yearly').removeClass('active');
                                $('.sign-up__cta .cta__quarterly').addClass('active');
                            }
                        }
                        $('.sign-up__cta .cta__price').text(((old_price * 12) - ((yearly_to_monthly - (yearly_to_monthly * .10)) * 12)).toFixed(2));
                        $('.sign-up__cta .extra-cta__text').removeClass('hide')
                        $('.sign-up__wrap').addClass('show__cta');
                        $('.sign-up__cta').removeClass('hide');
                        Cookies.set('affwp_ref', response.affiliate);
        
                        $('.applied_coupon').val(response.coupon)
                        
                        $('.sign-up__price').removeClass('coupon-50 coupon-10');
                        $('.sign-up__price').addClass('coupon-' + couponType);
                        $('.sign-up__form .affiliateCoupon').attr('disabled', true);

                        $this.parent().addClass('disabled');
                        $this.text('Coupon Applied');
                    } else {
                        $('.sign-up__form .affiliateCoupon').parent().addClass('error')
                        $('.sign-up__form .affiliateCoupon').parent().find('.input-error').text('Invalid Coupon').show();
                        $('.sign-up__form .affiliateCoupon').addClass('warning')
                    }

                    $this.removeClass('processing');
                    $this.show();
                    $this.parent().find('.spin-loader').hide();
                    $('.sign-up__price').show();
                    $('.sign-up__cta').show();
                    $('.sign-up__wrap .spin-loader').hide();
                }
            })

        }
    })
    $('body').on('submit', '.pricing__form:not(.blackfriday)', function (e) {
        e.preventDefault();
        var $btn = $('.cta__btn.coupon__button');
        let coupon = $('.coupon__container .coupon__input').val();
        $('.plans__old-price').css('visibility', 'visible');
        $('.noticeBox').removeClass('no-discount');

        var data = {
            'action': 'checkCoupon',
            'coupon': coupon,
        };

        $.ajax({
            url: book_beam_params.ajax_url,
            type: 'POST',
            data: data,
            beforeSend: function () {
                //Store price
                var prices_yearly = $('.plans__price .new_yearly_orig_price');
                var prices_monthly = $('.plans__price .old_monthly_orig_price');
                var prices_quarterly = $('.plans__price .new_quarterly_orig_price');
                if (data_prices_yearly.length === 0) {
                    prices_yearly.each(function (i, e) {
                        data_prices_yearly.push($(e).val());
                    })
                }
                if (data_prices_monthly.length === 0) {
                    prices_monthly.each(function (i, e) {
                        data_prices_monthly.push($(e).val());
                    })
                }
                if(data_prices_quarterly.length === 0){
                    prices_quarterly.each(function (i, e) {
                        data_prices_quarterly.push($(e).val());
                    })
                }
                $btn.find('div').show();
                $btn.attr('disabled', true);
            },
            success: function (res) {

                if (res.affiliate > 0) {
                    $('.pricing__switch li[data-pay=yearly]').click()
                    var msg = 'Coupon <b style="text-transform:uppercase;">' + res.coupon + '</b> applied.';
                    var couponType = res.couponType;
                    var quarterly_msg = ''
                    $('.handrwritten-text .text').text($('.handrwritten-text .text').attr('data-text-with-coupon'))
                    if($('.plans__cta.quarterly').length > 0){
                        quarterly_msg = ', up to 22% for quarterly';
                    }
                    if (couponType == 10) {
                        msg += ' Save up to 44% with annual' + quarterly_msg + ' or 15% for monthly for 6 months!';

                        var price = $('.plans__price .new_yearly_orig_price');
                        price.each(function (i, e) {
                            if (data_prices_yearly.length !== 0) {
                                // $(e).text(data_prices_yearly[i]);
                                $(e).val(data_prices_yearly[i]);
                            }
                        })
                        price.each(function (i, e) {
                            var price = $(e).val();
                            price = price.replace('/mo', '');
                            price = price.trim();

                            price = price.replace('$', '');
                            price = +price.trim()

                            var discount_price = price - (price * 0.1);
                            var discount_price_year = $($('.plans__price .new_yearly_orig_price')[i]).val();

                            discount_price_year = discount_price_year.replace('$', '');
                            discount_price_year = discount_price_year.replace('/mo', '');
                            discount_price_year = +discount_price_year.trim();

                            discount_price_year = discount_price_year - (discount_price_year * 0.1);

                            var original_price = $($('.plans__old-price')[i]).text();
                            original_price = original_price.replace('$', '');
                            original_price = original_price.replace('/mo', '');
                            original_price = +original_price.trim();

                            discount_price_year = discount_price_year * 12;
                            var saved_price = original_price * 12 - discount_price_year;
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');

                            $($('.plans__saved-price.annually')[i]).find('span').text('$' + saved_price.toFixed(2))

                            $($('.perks-list .perks-header .plans__saved-price.annually')[i]).find('span').text('$' + saved_price.toFixed(2))

                            $($('.perks-list .perks-header .plans__saved-price.annually')[i + 3]).find('span').text('$' + saved_price.toFixed(2))
                            
                            $(e).parent().find('.yearly').text('$' + discount_price.toFixed(2) + '/mo');
                            
                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .yearly:not(.plans__price-strike--through)').text('$' + discount_price.toFixed(2) + '/mo')

                            $($('.plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');

                            $($('.perks-list .plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');
                            $($('.perks-list .plans__cta:not(.quarterly)')[i + 3]).text('$' + discount_price_year.toFixed(2) + ' for the first year');

                        })
                        var price_50 = $('.plans__price .old_monthly_orig_price');
                        price_50.each(function (i, e) {
                            if (data_prices_monthly.length !== 0) {
                                // $(e).text(data_prices_monthly[i]);
                                $(e).val(data_prices_monthly[i]);
                            }
                        })
                        price_50.each(function (i, e) {
                            var price_50 = $(e).val();
                            price_50 = price_50.replace('/mo', '');
                            price_50 = price_50.trim();

                            price_50 = price_50.replace('$', '');
                            price_50 = +price_50.trim()

                            var discount_price = price_50 - (price_50 * 0.15);
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');
                            $($('.plans__saved-price.monthly')[i]).find('span').text('$' + ((price_50 * 6) * 0.15).toFixed(2))
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i]).find('span').text('$' + ((price_50 * 6) * 0.15).toFixed(2))
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i + 3]).find('span').text('$' + ((price_50 * 6) * 0.15).toFixed(2))
                            
                            $(e).parent().find('.monthly').text('$' + discount_price + '/mo');
                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .monthly:not(.plans__price-strike--through)').text('$' + discount_price + '/mo')
                            $('.plans__list').addClass('has-coupon')
                            $('.perks-list').addClass('has-coupon')

                        })
                        var price_quarterly = $('.plans__price .new_quarterly_orig_price');
                        price_quarterly.each(function (i, e) {
                            if (data_prices_quarterly.length !== 0) {
                                // $(e).text(data_prices_quarterly[i]);
                                $(e).val(data_prices_quarterly[i]);
                            }
                        })
                        var price_index = 2;
                        price_quarterly.each(function (i, e) {
                            var price_quarterly = $(e).val();
                            price_quarterly = price_quarterly.replace(' for 3 months', '');
                            price_quarterly = price_quarterly.trim();

                            price_quarterly = price_quarterly.replace('$', '');
                            price_quarterly = +price_quarterly.trim()
                            if(price_quarterly > 0){
                                var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');
                                var quarterly_discount = 20;
                                plan_title = plan_title.replace('publisher-', '').trim();
                                if(plan_title === 'pro'){
                                    quarterly_discount = 38;
                                }
                                var monthly_price =  $(e).parent().parent().find('.plans__old-price').text().replace('/mo', '');
                                monthly_price = monthly_price.trim();
    
                                monthly_price = monthly_price.replace('$', '');
                                monthly_price = +monthly_price.trim()

                                var saved_quarterly = $(e).parent().parent().find('.plans__saved_quarterly').val();
                                saved_quarterly = parseFloat(saved_quarterly);

                                var quarterly_orig_price = monthly_price * 3;
                                var discount_price = (monthly_price * 3) - (quarterly_discount + saved_quarterly);

                                $($('.plans__saved-price.quarterly')[i]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                               
                                $($('.plans__saved-price.quarterly')[i + 3]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                
                                $($('.perks-list .perks-header .plans__saved-price.quarterly')[i]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                $
                                ($('.perks-list .perks-header .plans__saved-price.quarterly')[i + 3]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                
                                $(e).parent().find('.quarterly').html('$' + discount_price + ' <span class="smaller">for 3 months</span>');
                                $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .quarterly:not(.plans__price-strike--through)').html('$' + discount_price + ' <span class="smaller">for 3 months</span>');
                                $('.plans__list').addClass('has-coupon')
                                $('.perks-list').addClass('has-coupon')
                                price_index += 1;
                            }
                        })
                        $('.plans__cta').removeClass('hidden');
                        $('.plans__cta.quarterly').addClass('hidden')
                        $('.plans__cta.quarterly').text('Billed every 3 months');
                    } else {
                        msg += ' Save up to 44% with annual'+ quarterly_msg +' or 15% for monthly for 6 months!';
                        //50 is monthly

                        $('.plans__old-price').css('visibility', 'visible');
                        //compute for total price
                        var price = $('.plans__price .old_monthly_orig_price');
                        price.each(function (i, e) {
                            if (data_prices_monthly.length !== 0) {
                                // $(e).text(data_prices_monthly[i]);
                                $(e).val(data_prices_monthly[i])
                            }
                        })
                        price.each(function (i, e) {
                            var price = $(e).val();
                            price = price.replace('/mo', '');
                            price = price.trim();

                            price = price.replace('$', '');
                            price = +price.trim()

                            var discount_price = price - (price * 0.15);
                            if (data_prices_yearly.length !== 0) {
                                var discount_price_year = data_prices_yearly[i];
                            } else {
                                var discount_price_year = $($('.plans__price .new_yearly_orig_price')[i]).val();
                            }
                            discount_price_year = discount_price_year.replace('$', '');
                            discount_price_year = discount_price_year.replace('/mo', '');
                            discount_price_year = +discount_price_year.trim();

                            discount_price_year = discount_price_year - (discount_price_year * 0.1);

                            //saved
                            discount_price_year = discount_price_year * 12;
                            var saved_price = price * 12 - discount_price_year;
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');

                            $($('.plans__saved-price.annually')[i]).find('span').text('$' + saved_price.toFixed(2))
                            
                            $($('.perks-list .plans__saved-price.annually')[i]).find('span').text('$' + saved_price.toFixed(2))

                            $($('.perks-list .perks-header .plans__saved-price.annually')[i]).find('span').text('$' + saved_price.toFixed(2))

                            $($('.perks-list .perks-header .plans__saved-price.annually')[i + 3]).find('span').text('$' + saved_price.toFixed(2))


                            $($('.plans__saved-price.monthly')[i]).find('span').text('$' + ((price * 6) * 0.15).toFixed(2))

                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i]).find('span').text('$' + ((price * 6) * 0.15).toFixed(2))
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i + 3]).find('span').text('$' + ((price * 6) * 0.15).toFixed(2))

                            $($('.plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');

                            $($('.perks-list .plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');
                            $($('.perks-list .plans__cta:not(.quarterly)')[i + 3]).text('$' + discount_price_year.toFixed(2) + ' for the first year');

                            $(e).parent().find('.monthly').text('$' + discount_price + '/mo');

                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .monthly:not(.plans__price-strike--through)').text('$' + discount_price + '/mo')

                            $('.plans__list').addClass('has-coupon')
                            $('.perks-list').addClass('has-coupon')

                        })
                        var price_10 = $('.plans__price .new_yearly_orig_price');
                        price_10.each(function (i, e) {
                            if (data_prices_yearly.length !== 0) {
                                // $(e).text(data_prices_yearly[i]);
                                $(e).val(data_prices_yearly[i]);
                            }
                        })
                        price_10.each(function (i, e) {
                            var price_10 = $(e).val();
                            price_10 = price_10.replace('/mo', '');
                            price_10 = price_10.trim();

                            price_10 = price_10.replace('$', '');
                            price_10 = +price_10.trim()


                            var discount_price = price_10 - (price_10 * 0.1);

                            var discount_price_year = $($('.plans__price .new_yearly_orig_price')[i]).val();

                            discount_price_year = discount_price_year.replace('$', '');
                            discount_price_year = discount_price_year.replace('/mo', '');
                            discount_price_year = +discount_price_year.trim();

                            discount_price_year = discount_price_year - (discount_price_year * 0.1);

                            var original_price = $($('.plans__old-price')[i]).text();
                            original_price = original_price.replace('$', '');
                            original_price = original_price.replace('/mo', '');
                            original_price = +original_price.trim();

                            discount_price_year = discount_price_year * 12;

                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');

                            $(e).parent().find('.yearly').text('$' + discount_price.toFixed(2) + '/mo');
                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .yearly:not(.plans__price-strike--through)').text('$' + discount_price.toFixed(2) + '/mo')

                            $($('.plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');

                            $($('.perks-list .plans__cta:not(.quarterly)')[i]).text('$' + discount_price_year.toFixed(2) + ' for the first year');
                            $($('.perks-list .plans__cta:not(.quarterly)')[i + 3]).text('$' + discount_price_year.toFixed(2) + ' for the first year');
                        })

                        var price_quarterly = $('.plans__price .new_quarterly_orig_price');
                        price_quarterly.each(function (i, e) {
                            if (data_prices_quarterly.length !== 0) {
                                // $(e).text(data_prices_quarterly[i]);
                                $(e).val(data_prices_quarterly[i]);
                            }
                        })
                        var price_index = 2;
                        price_quarterly.each(function (i, e) {
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');
                            var quarterly_discount = 20;
                            plan_title = plan_title.replace('publisher-', '').trim();
                            if(plan_title === 'pro'){
                                quarterly_discount = 38;
                            }
                            var price_quarterly = $(e).val();
                            price_quarterly = price_quarterly.replace(' for 3 months', '');
                            price_quarterly = price_quarterly.trim();

                            price_quarterly = price_quarterly.replace('$', '');
                            price_quarterly = +price_quarterly.trim()
                            if(price_quarterly > 0){
                                var monthly_price =  $(e).parent().parent().find('.plans__old-price').text().replace('/mo', '');
                                monthly_price = monthly_price.trim();
    
                                monthly_price = monthly_price.replace('$', '');
                                monthly_price = +monthly_price.trim()

                                var saved_quarterly = $(e).parent().parent().find('.plans__saved_quarterly').val()
                                saved_quarterly = parseFloat(saved_quarterly);

                                var quarterly_orig_price = monthly_price * 3;
                                var discount_price = (monthly_price * 3) - (quarterly_discount + saved_quarterly);

                                var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');

                                
                                $($('.plans__saved-price.quarterly')[i]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                $($('.perks-list .perks-header .plans__saved-price.quarterly')[i]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                
                                $($('.perks-list .perks-header .plans__saved-price.quarterly')[i + 3]).find('span').text('$' + (quarterly_orig_price- discount_price).toFixed(0))
                                
                                $(e).parent().find('.quarterly').html('$' + discount_price + ' <span class="smaller">for 3 months</span>');
                                $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .quarterly:not(.plans__price-strike--through)').html('$' + discount_price + ' <span class="smaller">for 3 months</span>')
                                $('.plans__list').addClass('has-coupon')
                                $('.perks-list').addClass('has-coupon')
                                price_index += 1;
                            }
                        })
                        $('.plans__cta.quarterly').addClass('hidden');
                        $('.plans__cta.quarterly').text('Billed every 3 months');
                        $('.plans__cta:not(.quarterly)').each(function(e){
                            $(this).parent().find('.plans__stored_cta').val($(this).text());
                            // $(this).text('15% Off for 6 Months');
                            if(couponType == 50){
                                $(this).removeClass('hidden')
                            }
                        })
                    }
                    Cookies.set('affwp_ref', res.affiliate);
                    $('.coupon_active').val(res.coupon);
                    $('.coupon_type').val(couponType);
                    $('.noticeBox').html(msg).show();
                } else {
                    $('.noticeBox').text('Invalid Coupon').show();
                }

                $btn.find('div').hide();
                $btn.attr('disabled', false);
            }
        })
        return false; // return false to prevent typical submit behavior
    });
    
    $('body').on('submit', '.pricing__form.blackfriday', function (e) {
        e.preventDefault();
        var $btn = $('.cta__btn.coupon__button');
        let coupon = $('.coupon__container .coupon__input').val();
        $('.plans__old-price').css('visibility', 'visible');
        $('.noticeBox').removeClass('no-discount');

        var data = {
            'action': 'checkCoupon',
            'coupon': coupon,
        };

        $.ajax({
            url: book_beam_params.ajax_url,
            type: 'POST',
            data: data,
            beforeSend: function () {
                //Store price
                var prices_monthly = $('.plans__price .old_monthly_orig_price');
                if (data_prices_monthly.length === 0) {
                    prices_monthly.each(function (i, e) {
                        data_prices_monthly.push($(e).val());
                    })
                }
                $btn.find('div').show();
                $btn.attr('disabled', true);
            },
            success: function (res) {

                if (res.affiliate > 0) {
                    $('.bf_pricing__switch li[data-pay=monthly]').click();
                    
                    var msg = 'Coupon ' + res.coupon + ' applied.';
                    var couponType = res.couponType;
                    if (couponType == 10) {
                        msg += ' Get 15% for monthly for 6 months!';

                        var price_50 = $('.plans__price .old_monthly_orig_price');
                        price_50.each(function (i, e) {
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');

                            var price_50 = $(e).val();
                            price_50 = price_50.replace('/mo', '');
                            price_50 = price_50.trim();

                            price_50 = price_50.replace('$', '');
                            price_50 = +price_50.trim()

                            var discount_price = price_50 - (price_50 * 0.15);
                            $($('.plans__list .plans__saved-price.monthly')[i]).find('span').text(((price_50 * 6) * 0.15).toFixed(2))
                            $($('.plans__list .plans__saved-price.monthly')[i]).addClass('show-save-text');
                            $($('.plans__list .plans__cta.monthly')[i]).addClass('show-plans-cta')
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i]).find('span').text(((price_50 * 6) * 0.15).toFixed(2))
                            $($('.perks-list .perks-header .plans__cta.monthly')[i]).addClass('show-plans-cta')
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i + 3]).find('span').text(((price_50 * 6) * 0.15).toFixed(2))
                            $($('.perks-list .perks-header .plans__cta.monthly')[i + 3]).addClass('show-plans-cta')
                            
                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .monthly:not(.plans__price-strike--through)').html('$' + discount_price + '<span style="font-size:14px;display:inline-block;margin-left:4px"> for 6 months</span>')
                            
                            $(e).parent().find('.monthly').html('$' + discount_price + ' <br/> <span style="font-size:22px;">for 6 months</span>');
                            $('.plans__list').addClass('has-coupon')
                            $('.perks-list').addClass('has-coupon')

                        })
                    } else {
                        msg += ' Get 15% for monthly for 6 months!';
                        //50 is monthly

                        $('.plans__old-price').css('visibility', 'visible');
                        //compute for total price
                        var price = $('.plans__price .old_monthly_orig_price');
                        $('.bf_pricing__switch li[data-pay=monthly]').click()
                        price.each(function (i, e) {
                            if (data_prices_monthly.length !== 0) {
                                // $(e).text(data_prices_monthly[i]);
                                $(e).val(data_prices_monthly[i])
                            }
                        })
                        price.each(function (i, e) {
                            var plan_title = $(e).parent().parent().find('h4').text().toLowerCase().replace(/\s+/g, '-');
                            
                            var price = $(e).val();
                            price = price.replace('/mo', '');
                            price = price.trim();

                            price = price.replace('$', '');
                            price = +price.trim()

                            var discount_price = price - (price * 0.15);
                            
                            $($('.plans__list .plans__saved-price.monthly')[i]).find('span').text(((price * 6) * 0.15).toFixed(2))
                            $($('.plans__list .plans__saved-price.monthly')[i]).addClass('show-save-text');
                            $($('.plans__list .plans__cta.monthly')[i]).addClass('show-plans-cta')

                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i]).find('span').text(((price * 6) * 0.15).toFixed(2))
                            $($('.plans__saved-price.monthly')[i]).addClass('show-save-text');
                            
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i + 3]).find('span').text(((price * 6) * 0.15).toFixed(2))
                            $($('.perks-list .perks-header .plans__saved-price.monthly')[i + 3]).addClass('show-save-text');

                            
                            $('.perks-list .perks-plan-' + plan_title.toLowerCase() + ' .perks-plan__price .monthly:not(.plans__price-strike--through)').html('$' + discount_price + '<span style="font-size:14px;display:inline-block;margin-left:4px"> for 6 months</span>')
                            
                            $(e).parent().find('.monthly').html('$' + discount_price + ' <br/> <span style="font-size:22px;">for 6 months</span>');

                            $('.plans__list').addClass('has-coupon')
                            $('.perks-list').addClass('has-coupon')

                        })
                    }
                    Cookies.set('affwp_ref', res.affiliate);
                    $('.coupon_active').val(res.coupon);
                    $('.coupon_type').val(couponType);
                    $('.noticeBox').html(msg).show();
                } else {
                    $('.noticeBox').text('Invalid Coupon').show();
                }

                $btn.find('div').hide();
                $btn.attr('disabled', false);
            }
        })
        return false; // return false to prevent typical submit behavior
    });
    $('#affwp-login-form').on('submit', function (e) {
        var user_login = $('#affwp-login-user-login').val();
        var password = $('#affwp-login-user-pass').val();
        $('.affwp-custom-error-container').children().remove();
        $('#affwp-login-user-login').removeClass('warning');
        $('#affwp-login-user-pass').removeClass('warning');

        if (user_login == '') {
            e.preventDefault();
            $('#affwp-login-user-login').addClass('warning');
            var html = '';
            var errormessage = '<p class="affwp-error">Please enter username</p>'
            if ($('.affwp-custom-error-container.login .affwp-errors').length > 0) {
                $('.affwp-custom-error-container.login .affwp-errors').append(errormessage)
            } else {
                html += '<div class="affwp-errors">'
                html += errormessage;
                html += '</div>';
                $('.affwp-custom-error-container.login').html(html)
            }

        }
        if (password == '') {
            e.preventDefault();
            $('#affwp-login-user-pass').addClass('warning');
            var html = '';
            var errormessage = '<p class="affwp-error">Please enter password</p>'
            if ($('.affwp-custom-error-container.login .affwp-errors').length > 0) {
                $('.affwp-custom-error-container.login .affwp-errors').append(errormessage)
            } else {
                html += '<div class="affwp-errors">'
                html += errormessage;
                html += '</div>';
                $('.affwp-custom-error-container.login').html(html)
            }

        }
    })
    $('#affwp-login-form input').on('input propertychange paste', function (e) {
        if ($(this).val() != '') {
            $(this).removeClass('warning');
            $('.affwp-custom-error-container').children().remove();
        }
    })
    $('#affwp-register-form').on('submit', function (e) {
        $('#affwp-register-form input, #affwp-register-form textarea').each(function (i, e) {

            if ($(e).attr('required') && $(e).val() == '') {
                e.preventDefault();
                $(e).addClass('warning');
            }
        })
    })
    $('#affwp-register-form input').on('input propertychange paste', function (e) {
        if ($(this).val() != '') {
            $(this).removeClass('warning');
            $('.affwp-custom-error-container').children().remove();
        }
    })

    $('body').on('click', '.sign-up__switcher-btn', function (e) {
        var $this = $(this);
        var switcher = $this.attr('data-signup-switcher');

        $('.plan_period').val(switcher)
        if (switcher == 'annual') {
            $('.sign-up__switcher').removeClass('monthly');
            $('.sign-up__switcher').addClass('annual');
            // $('.monthly-orig-price').hide();
        } else {
            $('.sign-up__switcher').addClass('monthly');
            $('.sign-up__switcher').removeClass('annual');
        }
        //first letter capital
        var firstLetter = switcher.charAt(0).toUpperCase();
        var rest = switcher.slice(1);
        switcher = firstLetter + rest;

        $('.sign-up__plan-title span').text(switcher);
    })
    var lastScrollTop = 0;
    $(window).on('scroll', function (e) {
        $('.off-canvas-wrapper-inner').removeClass('hidden');
        //Check if blog-sidebar__social is on very top or bottom of its container

        if($('.blog-sidebar__social').length > 0){
            var st = $(this).scrollTop();

            var $sidebar = $('.blog-sidebar__social');
            var $blog_sidebar = $('.blog-sidebar');
            var $toc_sidebar = $('.toc__sidebar');
            var $sidebar_container = $('.single-blog-article');
            var $sidebar_container_height = $sidebar_container.height();
            var $sidebar_height = $sidebar.height();
            var $sidebar_top = $blog_sidebar.offset().top;
            var $sidebar_bottom = $sidebar_top + $sidebar_height;


            //scroll direction
            var scroll_direction = 'down';
            if (st  < lastScrollTop){
                scroll_direction = 'up';
            }
            lastScrollTop = st;

            
            if(scroll_direction == 'up'){
                if($sidebar_top < $sidebar_container.offset().top + $sidebar_height + 300){
                    $sidebar.fadeOut(200);
                    if($toc_sidebar.length > 0){
                        $toc_sidebar.fadeOut(200);
                    }
                }
                if($sidebar_top < $sidebar_container_height + $sidebar_height && $sidebar_top >= $sidebar_container.offset().top + $sidebar_height + 300 ){
                    if($toc_sidebar.length > 0){
                        $toc_sidebar.fadeIn(200);
                    }
                    $sidebar.fadeIn(200);
                }
            }
            
            if($('body').hasClass('single-bz_changelogs') || $('body').hasClass('single-changelog')){
                var minus = 750
            }else{
                var minus = 500;
            }
            if(scroll_direction == 'down'){
                if($sidebar_top > $sidebar_container_height - minus ){
                    $sidebar.fadeOut(200);
                    if($toc_sidebar.length > 0){
                        $toc_sidebar.fadeOut(200);
                    }
                }
                if($sidebar_top < $sidebar_container_height - minus){
                    $sidebar.fadeIn(200)
                    if($toc_sidebar.length > 0){
                        $toc_sidebar.fadeIn(200);
                    }
                }
            }

        }
    })
    //When the user click on the sign up button scroll to the id
    $('body').on('click', '.login-menu .signup', function (e) {
        e.preventDefault();
        $(this).addClass('scroll-to-pricing');
    })
    $('body').on('click', '.docs__link__footer, .support__link__header , .docs__link__header, .cs__link__footer, .affiliate-signup .sign-up__login.override, .affiliate-signup .affwp-lost-password, .homepage__link__header', function (e) {
        e.preventDefault();
        var href = $(this).find('a').attr('href');

        window.location.href = href
    })
    $('body').on('click', '.apply-form-link' , function(e){
        e.preventDefault()
        var href = $(this).attr('data-link');
        window.location.href = href;
    })
    $('body').on('click', '#betterdocs-breadcrumb li', function(e){
        // IMPORTANT to Remove animate Once every time page is loaded for one time animation
        sessionStorage.removeItem('animate_once_reveal_scale');
        sessionStorage.removeItem('animate_once_reveal_split');
        sessionStorage.removeItem('animate_once_reveal_text_left');
        sessionStorage.removeItem('animate_once_reveal_top');
        sessionStorage.removeItem('animate_once_reveal_simple');
    })
    $('body').on('focusin', '.betterdocs-page .betterdocs-search-field', function (e) {
        $('.betterdocs-page .betterdocs-searchform').addClass('focus');
    })

    $('body').on('focusout', '.betterdocs-page .betterdocs-search-field', function (e) {
        $('.betterdocs-page .betterdocs-searchform').removeClass('focus');
    })
    // $('body').on('mouseenter', '.footer__menu .features__footer a', function (e) {
    //     $('.features__header .dropdown.vertical').addClass('active')
    // })
    $('body').on('click', '.contact-page .wpcf7-submit, .apply-form-page .wpcf7-submit' , function () {

        var text = $(this).val();
        var $this = $(this);
        $(this).val('Submitting...');
        document.addEventListener('wpcf7submit', function (event) {
            $this.val(text);
        }, false);
    })

    $('body').on('mouseenter', '.resources-header__menu, .features__header', function (e) {
        $('.features__header [data-toggle] .menu-item-has-children').removeClass('feature-active'); 
        $(this).find('.dropdown.vertical[data-toggle]').addClass('active');
        $(this).find('[data-toggle] .menu-item-has-children:first-child .dropdown.vertical[data-toggle-nested]').addClass('active')
        $(this).find('[data-toggle] .menu-item-has-children:first-child').addClass('feature-active')
    })
    $('body').on('mouseenter', '.features__header [data-toggle] .menu-item-has-children', function(e){
        $('.features__header [data-toggle] .menu-item-has-children').removeClass('feature-active'); 
        $('.features__header [data-toggle] .menu-item-has-children .dropdown.vertical[data-toggle-nested]').removeClass('active');
        $(this).find('.dropdown.vertical[data-toggle-nested]').addClass('active');
        $(this).addClass('feature-active')
    });
    $('body').on('mouseleave', '.resources-header__menu, .features__header', function (e) {
        $(this).find('.dropdown.vertical').removeClass('active');
    })
    $('body').on('click' , '#menu-main-nav-1 li a, .footer li a', function(e){
        $('#usercom-widget, .crisp-client').fadeOut(200);
    })
    $(window).on('scroll', function(e){
        if ($(window).scrollTop() > 100) {
            if($('body').is('.home,.front,.pricing:not(.about,.thankyou),.page-template-pricing, .betterdocs-page, .single-docs, .post-type-archive-docs, .tax-doc_category')){
                $('#usercom-widget, .crisp-client').fadeIn(200);
            }else{
                $('#usercom-widget, .crisp-client').fadeOut(200);
            }
        }
    })

    //Calculator tabs animation
    $('body').on('click', '.pr-tabs li a', function(){
        $(this).parent().parent().find('.selector').addClass('animating')
        $('.pr-tabs li').removeClass("active animated");
        // var activeWidth = $(this).parent().innerWidth();
        // var itemPos = $(this).parent().position();
        $(this).parent().addClass('active animated')
        // $(this).parent().parent().find('.selector').css({
        //     "visibility": "visible",
        //     "left":itemPos.left + "px", 
        //     "top": css_top + 'px',
        //     "width": activeWidth + "px",
        // })

    })
    $('body').on('click', '#menu-resources.footer__menu li a, #menu-main-nav-1 li a', function(e){
        var href = $(this).attr('href');
        if($('.pr-tabs').length > 0 ){
            $(".selector").addClass('animating')

            $('.pr-tabs li').removeClass("active");

            $('.pr-tabs li a').each(function(i,e){
                //check if text contains
                if($(e).attr('href') == href){
                    var activeWidth = $(e).parent().innerWidth();
                    var itemPos = $(e).parent().position();
                    $(e).parent().addClass('active')
                    $(".selector").css({
                        "visibility": "visible",
                        "left":itemPos.left + "px", 
                        "width": activeWidth + "px",
                    })
                }
            })
        }
    })
    $('body').on('click', '.footer li a, #menu-main-nav-1 li a', function(e){
        // IMPORTANT to Remove animate Once every time page is loaded for one time animation
        sessionStorage.removeItem('animate_once_reveal_scale');
        sessionStorage.removeItem('animate_once_reveal_split');
        sessionStorage.removeItem('animate_once_reveal_text_left');
        sessionStorage.removeItem('animate_once_reveal_top');
        sessionStorage.removeItem('animate_once_reveal_simple');
    })
    $('body').on('click', '#menu-main-nav-1 li a', function(e){
        $('#menu-main-nav-1 li').removeClass('is-active is-active-1')

    })
    $('body').on('click', '#menu-main-nav-1 li.menu-item-has-children a', function(e){
        e.preventDefault();
        if($(this).attr('href') == '#'){
            return;
        }
        $('#menu-main-nav-1 li').removeClass('is-active is-active-1')
        $(this).parent().parent().parent().addClass('is-active is-active-1');
        $('#menu-main-nav-1 li.menu-item-has-children .dropdown li').removeClass('current-menu-item');
        $(this).parent().addClass('current-menu-item');
    })

    $('.apply-form-page .apply-form-content form .wpcf7-validates-as-required').on('blur focusout', function(e){
        if($(this).val() != ''){
            $(this).removeClass('wpcf7-not-valid');
            $(this).parent().find('.wpcf7-not-valid-tip').remove()
        }
    })
    
    // if($('#kdp-cat__browser #kdp-cat__dt').length > 0){
    //     $('#kdp-cat__browser #kdp-cat__dt').DataTable({
    //         "bLengthChange": false,
    //         "bFilter": true,
    //         "bInfo": false,
    //         "searching": false,
    //         "columnDefs": [
    //             { "orderable": false, "targets": 0 }
    //         ]
    //     })
    // }
    $('body').on('click','.calculator-container#kdp-cat__browser .category__title .table__icon', function(e){
        if(!$(this).parent().hasClass('animated-background')){
            copyText($(this).parent().find('a'), $(this));
        }
    })
    $('body').on('click', '.calculator-container#kdp-keyword__generator .category__title .open__book', function(e){
        if($(this).parent().hasClass('animated-background')){
            return false;
        }
    })
    $('body').on('click','.calculator-container#kdp-keyword__generator .category__title .table__icon', function(e){
        if(!$(this).parent().hasClass('animated-background')){
            copyText($(this).parent().find('.category__text'), $(this));
        }
    })

    $('body').on('click', '.limit__popup .limit__popup_close_container .limit__popup_close', function(e){
        e.preventDefault();
        $('.limit__popup').removeClass('active')
    })
    $('body').on('click', '.limit__popup .limit__popup_overlay', function(e){
        e.preventDefault();
        $('.limit__popup').removeClass('active')
    })
    
    $('body').on('click', '.ce_lite_popup .ce_lite_popup_close_container .ce_lite_popup_close', function(e){
        e.preventDefault();
        $('.ce_lite_popup').removeClass('active')
    })
    $('body').on('click', '.ce_lite_popup .close_ce_lite_popup', function(e){
        //if href is # then dont do anything
        if($(this).attr('href') == '#'){
            e.preventDefault();
        }
        $('.ce_lite_popup').removeClass('active')
    })

    $('body').on('click', '.affiliate_popup .affiliate_popup_button', function(e){
        $('.affiliate_popup').removeClass('active')
    })
    $('body').on('click', '.affiliate_popup .affiliate_popup_close_container .affiliate_popup_close', function(e){
        e.preventDefault();
        $('.affiliate_popup').removeClass('active')
    })
    $('body').on('click', '.affiliate_popup .close_affiliate_popup', function(e){
        //if href is # then dont do anything
        if($(this).attr('href') == '#'){
            e.preventDefault();
        }
        $('.affiliate_popup').removeClass('active')
    })
    $('body').on('click', '.affiliate_popup .affiliate_popup_button', function(e){
        $('.affiliate_popup').removeClass('active')
    })
    
    $('body').on('click', '.affiliate_popup .affiliate_popup_overlay', function(e){
        e.preventDefault();
        $('.affiliate_popup').removeClass('active')
    })

    $('body').on('click', '.answers__contents .answers__item', function(e){
        $('.answers__contents .answers__item').removeClass('selected');
        $(this).addClass('selected');
        $('.answers__input').hide();
        if($(this).hasClass('others')){
            $('.answers__input').show();
        }else{ 
            if(!$(this).hasClass('processing')){
                $(this).addClass('processing')
                submitSurveyAns($(this).text());
            }
        }
        
    })

    function submitSurveyAns(text){

        $.ajax({
            url: book_beam_params.ajax_url,
            method: "POST",
            data: {
                action: 'uninstall_survey',
                nonce: book_beam_params.nonce,
                text:text,
            },
            beforeSend: function(e){

            },
            success: function(response){
                $('.answers__contents .answers__item, .answers__input textarea, .answers__input .answers__input-btn .button').removeClass('processing');
            }
        })
        $('.feat-hero__second_title').fadeOut('slow', function(){
            $('.after_survey__container .feat-hero__second_title').text('Thank you!').fadeIn('slow')
        });
        $('.answers').fadeOut('slow', function(){
            
            $('.after_survey__container').fadeIn('slow');
            $(this).remove();
        });
    }
    $('body').on('input', '.answers__input textarea', function(){
        var value = $(this).val();
        if(value != ''){
            $('.answers__input .answers__input-btn').show()
        }else{
            $('.answers__input .answers__input-btn').hide()
        }
    })
    
    $('body').on('click', '.answers__input .answers__input-btn .button', function(e){
        e.preventDefault();
        var value = $('.answers__input textarea').val();
        if(!$(this).hasClass('processing') && value != ''){
            $(this).addClass('processing')
            submitSurveyAns(value);
        }
    })
    $('body').on('click', '.pricing-v2 .perks-header .button', function (e) {
        e.preventDefault();

        var $this = $(this);

        var plan = $this.attr('data-plan');

        var plan_list = $('.plans__list li[data-plan-list="' + plan + '"]');
        var pricing = $('.pricing__switch').length > 0 ? $('.pricing__switch').attr('data-pay-switch') || 'annual' : $('.bf_pricing__switch').attr('data-pay-switch') || 'annual';
        var plan_title = plan_list.find('h4').text();
        var monthly_price = plan_list.find('.plans__price .monthly').text();
        var yearly_price = plan_list.find('.plans__price .yearly').text();
        var old_price = plan_list.find('.plans__old-price').text();
        var quarterly_price = plan_list.find('.plans__price .quarterly') ? plan_list.find('.plans__price .quarterly').text() : '';
        var original_monthly_d_price = yearly_price;
        var original_yearly_price = 0;
        var yearly_del_text = plan_list.find('.plans__price .yearly del').text();
        var quarterly_del_text = plan_list.find('.plans__price .quarterly') ? plan_list.find('.plans__price .quarterly del').text() : '';
        var quarterly_saved = $(plan_list.find('.plans__saved-price.quarterly div')[plan_list.find('.plans__saved-price.quarterly div').length - 1]).text();
        var monthly_saved = $(plan_list.find('.plans__saved-price.monthly div')[plan_list.find('.plans__saved-price.monthly div').length - 1]).text();
        var yearly_saved = $(plan_list.find('.plans__saved-price.annually div')[plan_list.find('.plans__saved-price.annually div').length - 1]).text();
        
        if (pricing == 'yearly') {
            pricing = 'annual';
        }

        monthly_price = monthly_price.replace('/mo', '');
        yearly_price = yearly_price.replace('/mo', '');
        old_price = old_price.replace('/mo', '');

        yearly_price = yearly_price.replace(/\$/g, '');
        yearly_del_text = yearly_del_text.replace('$', '');
        monthly_price = monthly_price.replace('$', '');
        old_price = old_price.replace('$', '');
        if($this.parent().parent().hasClass('blackfriday')){
            var yearlyElement = plan_list.find('.plans__price .yearly');
            const clone = yearlyElement.clone();
            clone.find('del').remove();
            yearly_price = clone.text();
            yearly_price = yearly_price.replace('/mo', '');
            yearly_price = yearly_price.replace(/\$/g, '');

            yearly_price = yearly_price.replace('for the first year', '');
            yearly_price = yearly_price.replace(yearly_del_text, '');
            monthly_price = monthly_price.replace('for the first 6 months', '');
            yearly_price = (parseFloat(yearly_price) / 12).toFixed(2);
            original_monthly_d_price = '$' + yearly_price + '/mo';
            var prices_yearly = $('.plans__price .yearly');
            var prices_monthly = $('.plans__price .monthly');

            if(quarterly_price != ''){
                quarterly_price = quarterly_price.replace(quarterly_del_text, '');
            }

            if (data_prices_yearly.length === 0) {
                prices_yearly.each(function (i, e) {
                    data_prices_yearly.push($(e).text());
                })
            }
            if (data_prices_monthly.length === 0) {
                prices_monthly.each(function (i, e) {
                    var price = $(e).text();
                    price = price.replace('/mo', '');
                    price = price.replace('$', '');
                    price = parseFloat(price) * 1;
                    price = '$' + price + '/mo';
                    data_prices_monthly.push(price);
                })
            }

        }
        
        monthly_price = +monthly_price.trim();
        yearly_price = +yearly_price.trim();
        old_price = old_price.trim();
        var plan_type = '';
        if (plan_title == 'Starter') {
            plan_type = 'Starter';
        }
        if (plan_title == 'Basic') {
            plan_type = 'basic';
        }
        if (plan_title == 'Pro') {
            plan_type = 'pro';
        }
        if(plan_title == 'Publisher Pro') {
            plan_type = 'publisherpro';
        }
        if(plan_title == 'Publisher'){
            plan_type = 'publisher';
        }

        if (pricing == 'annual') {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[plan_list.index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + original_yearly_price.toFixed(2);
            old_price = '$' + old_price + '/mo';
            monthly_price = '$' + monthly_price;

        } else {
            if (data_prices_monthly.length > 0) {
                original_yearly_price = data_prices_monthly[plan_list.index()];
                original_yearly_price = original_yearly_price.replace('$', '');
                original_yearly_price = original_yearly_price.replace('/mo', '');
                original_yearly_price = +original_yearly_price.trim();
                original_yearly_price = original_yearly_price * 12;
            } else {
                original_yearly_price = old_price * 12;
            }
            original_yearly_price = '$' + original_yearly_price.toFixed(2);
            old_price = '$' + old_price + '/mo';

            monthly_price = '$' + monthly_price;
        }

        yearly_price = yearly_price * 12;
        yearly_price = parseFloat(yearly_price.toFixed(2));
        if(yearly_price % 1 !== 0){
            yearly_price = '$' + yearly_price.toFixed(2);
        }else{
            yearly_price = '$' + yearly_price;
        }
        
        if (!$this.hasClass('processing')) {
            $('.perks-list .perks-header .button').addClass('processing');
            $('.plans__info .button').addClass('processing');
            var isQuarterly = 0;
            
            if($('main').hasClass('pricing-v2') && plan_title != 'Starter' && pricing != 'monthly'){
                isQuarterly = 1;
            }
            var data = {
                plan_type: plan_type,
                pricing: pricing,
                plan_title: plan_title,
                monthly_price: monthly_price,
                original_yearly_price: original_yearly_price,
                original_monthly_d_price: original_monthly_d_price,
                original_quarterly_price: quarterly_del_text,
                yearly_price: yearly_price,
                old_price: old_price,
                is_quarterly: isQuarterly,
                quarterly_price: quarterly_price,
                quarterly_saved: quarterly_saved,
                monthly_saved: monthly_saved,
                yearly_saved: yearly_saved,
                coupon: $('.coupon_active').val() != '' ? $('.coupon_active').val() : '',
                coupon_type: $('.coupon_type').val() != '' ? $('.coupon_type').val() : '',
            }
            if($this.hasClass('bf-button')){
                Cookies.set('bookbeam_pricing_data_bf', JSON.stringify(data), { expires: 1, path: '/' });
            }else{
                Cookies.set('bookbeam_pricing_data', JSON.stringify(data), { expires: 30, path: '/' });
            }
            window.dataLayer.push({
              'event' : 'PricingToRegistrationClick',
              'eventCategory' : 'Button',
              'eventAction' : 'Click'
            });

            
            $this.find('span').show();
            $this.find('.sk-circle').hide();
            $('.plans__info .button').removeClass('processing');
            $('.perks-list .perks-header .button').removeClass('processing');

            // if($('body').hasClass('early-bird-special')){
            //     window.location.href = site_url + '/early-bird-special-registration/';
            //     return;
            // }
            // if (window.location.pathname.includes('black-friday')) {
            //     // window.location.href = site_url + '/registration';
            //     window.location.href = site_url + '/black-friday-registration/';
            //     return;
            // }else{
            //     window.location.href = site_url + '/registration';
            //     return;
            //     // window.location.href = site_url + '/registration-wp-stripe-form';
            // }

            let bbParameter = '?plan=' + plan_title.replace(' ', '') + ' ' + pricing;
            if(Cookies.get('affwp_ref')){
                bbParameter += '&affiliate=' + Cookies.get('affwp_ref');
            }
            if($('.coupon_active').val() != ""){
                bbParameter += '&coupon=' + $('.coupon_active').val();
            }
            let url = "https://app.zB2mCiISUVux.io/#/auth/signup" //https://qa.zB2mCiISUVux.io/#/auth/signup
            window.location.href = url + bbParameter;
        }

    })
    // $('body').on('click', '.ce_lite_popup .ce_lite_overlay', function(e){
    //     e.preventDefault();
    //     $('.ce_lite_popup').removeClass('active')
    // })
    // if($('.ce_lite_popup').length > 0){
    //     setTimeout(function(e){
    //         $('.ce_lite_popup').addClass('active')
    //     }, 1500)
    // }
    function copyText(element, button){
        var el = $(element).text();
        var tmp = $("<input>");
        
        $(element).append(tmp);
        tmp.val(el).select();
        document.execCommand("copy");
        tmp.remove();

        $(button).html('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#2ec771" width="100%" height="100%" fit="" preserveAspectRatio="xMidYMid meet" focusable="false"><path d="M0 0h24v24H0z" fill="none"></path><path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z"></path></svg>')
        setTimeout(function(e){
            $(button).html('<svg xmlns="http://www.w3.org/2000/svg" fit="" height="100%" width="100%" preserveAspectRatio="xMidYMid meet" focusable="false"><path d="M0 0h24v24H0z" fill="none"></path><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"></path></svg>')
        }, 500)
    }
    function navigation() {
        const menuParent = document.querySelector('#menu-main-nav-1 .menu-item-has-children'),
        menuChild = menuParent.querySelector('[data-toggle]'),
        menuParentMobA = $('.mobile-off-canvas-menu #menu-main-nav .menu-item-has-children a[parent-anchor="true"]'),
        menuParentMob = $('.mobile-off-canvas-menu #menu-main-nav .menu-item-has-children a[parent-anchor="true"]'),
        menuChildMob = $(menuParentMob).find('.nested.menu'),
        menuMob = document.querySelector('.mobile-off-canvas-menu #menu-main-nav');

        menuParent.addEventListener('mouseenter', () => {
          menuChild.classList.add('active');
        })
      
        menuParent.addEventListener('mouseleave', () => {
          menuChild.classList.remove('active');
        })
        //get main menu top style
        var menuTop = $('#menu-main-nav').outerHeight();
        
        menuParentMobA.on('click', function(e) {
            var id = $(this).parent().attr('id')
            var $this = $(this).parent();
            var depthClassPattern = /menu-item-depth-(\d+)/;
            var allClasses = $this.attr('class').split(' ');
            var depthIndex = null;
            $.each(allClasses, function(index, className) {
                var match = className.match(depthClassPattern);
                if (match) {
                    depthIndex = match[1]; // Capture the depth index
                }
            });
            menuParentMob.each(function(i,e){
                // if($(this).attr('id') != id){
                    $(e).removeClass('dropdown-active')
                    $(this).find('.nested.menu').slideUp(200).removeClass('active');
                // }
            })

            if(!$this.hasClass('dropdown-active')){
                $this.addClass('dropdown-active')
                $this.find('.vertical.nested.menu.nested-depth-' + depthIndex).slideDown(200,function(){
                    if($this.parent().hasClass('resources-header__menu')){
                        $this.attr('style', 'display:flex')
                    }
                }).addClass('active');
            }else{
                $this.removeClass('dropdown-active')
                $this.find('.vertical.nested.menu.nested-depth-' + depthIndex).slideUp(200, function(){
                    if($this.parent().hasClass('resources-header__menu')){
                        $this.removeAttr('style')
                    }
                }).removeClass('active')
                // $('.mobile-off-canvas-menu').removeAttr('style')

            }
            
        })
      
        
      
        const menuBtn = document.querySelector('[data-mob-menu]'),
          body = document.querySelector('body'),
          offCanvas = document.querySelector('[data-offcanvas]');
      
        if (menuBtn) {
          menuBtn.addEventListener('click', () => {
            menuBtn.classList.toggle('active');
            body.classList.toggle('menu-active');
            offCanvas.classList.toggle('active');
          })
        }
    }
    navigation();
    bannerCountdownInit();
    
    $('body').on('click', '.plans__list .see-features', function(e){
        e.preventDefault();
        $('html, body').animate({
            scrollTop: $('.perks-list').offset().top
        }, 1000); 
    })
    $('body').on('click', '.special-offer-banner .close-banner', function(e){
        e.preventDefault();
        // $('.special-offer-banner').toggleClass('show');
        $('.special-offer-banner').removeClass('show');
        // if($('.special-offer-banner').hasClass('show')){
        //     $(this).html('Show banner')
        //     $(this).html('<i class="fa fa-times"></i>')
        // }else{
        //     $(this).html('<i class="fa fa-arrow-up"></i>')
        // }
    })
    function bannerCountdownInit(){
        if($('.special-offer-banner').length > 0){
            var date = $('.special-banner-countdown').attr('data-expires-in');
            var dateTime = date
            var countDownDate = date * 1000;
            
            var x = setInterval(function() {
                var now = new Date().getTime();
                var distance = countDownDate - now;
                var days = Math.floor(distance / (1000 * 60 * 60 * 24));
                var hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
                var minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
                var seconds = Math.floor((distance % (1000 * 60)) / 1000);
                if(days < 10){
                    days = '0' + days
                }
                if(hours < 10){
                    hours = '0' + hours
                }
                if(minutes < 10){
                    minutes = '0' + minutes
                }
                if(seconds < 10){
                    seconds = '0' + seconds
                }

                $('.special-banner-countdown .days .timer').html(days)
                $('.special-banner-countdown .hours .timer').html(hours)
                $('.special-banner-countdown .minutes .timer').html(minutes)
                $('.special-banner-countdown .seconds .timer').html(seconds)
                if (distance < 0) {
                    clearInterval(x);
                }
            })
        }
    }
});
